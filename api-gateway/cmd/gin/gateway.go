package main

import (
	"fmt"
	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-gonic/gin"
	"github.com/slink-go/api-gateway/cmd/common/templates"
	"github.com/slink-go/api-gateway/gateway"
	gwctx "github.com/slink-go/api-gateway/middleware/context"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/logging"
	"net/http"
	"net/url"
)

type GinBasedGateway struct {
	logger          logging.Logger
	contextProvider gwctx.Provider
	reverseProxy    *proxy.ReverseProxy
	registry        registry.ServiceRegistry
	quitChn         chan struct{}
}

//region - initializers

func NewGinBasedGateway() gateway.Gateway {
	gw := GinBasedGateway{
		logger:          logging.GetLogger("gin-gateway"),
		contextProvider: gwctx.CreateContextProvider(),
	}
	return &gw
}
func (g *GinBasedGateway) WithAuthProvider(ap security.AuthProvider) gateway.Gateway {
	g.contextProvider.WithAuthProvider(ap)
	return g
}
func (g *GinBasedGateway) WithUserDetailsProvider(udp security.UserDetailsProvider) gateway.Gateway {
	g.contextProvider.WithUserDetailsProvider(udp)
	return g
}
func (g *GinBasedGateway) WithReverseProxy(reverseProxy *proxy.ReverseProxy) gateway.Gateway {
	g.reverseProxy = reverseProxy
	return g
}

func (g *GinBasedGateway) WithRateLimiter(limit rate.Limiter) gateway.Gateway {
	//limiterConfig := limiter.Config{
	//	Max:                    limit.GetLimit(),
	//	Expiration:             time.Minute,
	//	SkipFailedRequests:     false,
	//	SkipSuccessfulRequests: false,
	//	LimiterMiddleware:      limiter.FixedWindow{},
	//	LimitReached: func(c *fiber.Ctx) error {
	//		return c.Status(http.StatusTooManyRequests).Send([]byte(fmt.Sprintf("Too Many Requests\n")))
	//	},
	//	KeyGenerator: func(c *fiber.Ctx) string {
	//		proxyIP := c.Get("X-Forwarded-For")
	//		if proxyIP != "" {
	//			return proxyIP
	//		} else {
	//			return c.IP()
	//		}
	//	},
	//}
	//g.limiter = limiter.New(limiterConfig)
	return g
}
func (g *GinBasedGateway) WithRegistry(registry registry.ServiceRegistry) gateway.Gateway {
	g.registry = registry
	return g
}
func (g *GinBasedGateway) WithQuitChn(chn chan struct{}) gateway.Gateway {
	g.quitChn = chn
	return g
}

func (g *GinBasedGateway) Serve(addresses ...string) {

	if addresses == nil || len(addresses) == 0 {
		panic("service address(es) not set")
	}

	if len(addresses) > 1 && addresses[1] != "" {
		go NewService("monitor").
			//WithPrometheus().
			//WithHandler("/monitor", monitor.New(monitor.Config{Title: "VOID API Gateway (monitoring)"})) // TODO: fiber-like monitoring
			WithGetHandlers("/", g.monitoringPage).
			WithGetHandlers("/list", g.listRemotes).
			WithStatic("/s", "./static").
			Run(addresses[1])
	}
	if addresses[0] != "" {
		NewService("proxy").
			WithPrometheus().
			WithMiddleware(headersCleanupMiddleware()). // cleanup incoming headers
			//WithMiddleware(csrf.New()). <-- implement it for Gin
			WithMiddleware(helmet.Default()).
			//WithMiddleware(timeoutMiddleware(100 * time.Millisecond)).
			WithMiddleware(proxyTargetResolverMiddleware(g.reverseProxy)).
			WithMiddleware(contextMiddleware(g.contextProvider)).
			WithNoRouteHandler(g.proxyHandler).
			WithQuitChn(g.quitChn).
			Run(addresses[0])

		//func (g *GinBasedGateway) setupProxyMiddleware() {
		// logger | должен быть первым, чтобы фиксировать отлупы от другого middleware и latency запросов
		// кстати о latency - нужна metricMiddleware
		//g.service.Use(logger.New())
		//g.engine.Use(g.customLoggingMiddleware)
		// endregion

	} else {
		panic("service port not set")
	}
}

// endregion
// region - monitoring

func (g *GinBasedGateway) monitoringPage(ctx *gin.Context) {
	ctx.Set("Content-Type", "text/html")
	t := templates.ServicesPage(templates.Cards(g.registry.List()))
	if err := t.Render(ctx.Request.Context(), ctx.Writer); err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}
func (g *GinBasedGateway) listRemotes(ctx *gin.Context) {
	if g.registry != nil {
		data := g.registry.List()
		if data == nil || len(data) == 0 {
			ctx.AbortWithStatus(http.StatusNoContent)
		} else {
			ctx.IndentedJSON(http.StatusOK, data)
		}
	} else {
		ctx.AbortWithStatus(http.StatusNoContent)
	}
}

// endregion
// region - proxy

func (g *GinBasedGateway) proxyHandler(ctx *gin.Context) {

	//defer func() {
	//	// а как вернуть RequestScoped-ошибку?
	//	if err := recover(); err != nil {
	//		g.logger.Warning("panic: %v", err)
	//	}
	//}()

	proxyTarget, statusCode, err := g.getProxyTarget(ctx)
	if err != nil {
		ctx.AbortWithError(statusCode, err)
	}

	g.logger.Trace("proxying %s", proxyTarget)

	ctx.Set("X-Forwarded-For", ctx.RemoteIP())
	ctx.Set("X-Real-Ip", ctx.ClientIP())

	g.reverseProxy.Proxy(ctx, proxyTarget).ServeHTTP(ctx.Writer, ctx.Request)

}

// endregion
// region - common

func (g *GinBasedGateway) getProxyTarget(ctx *gin.Context) (*url.URL, int, error) {
	value, ok := ctx.Get(gwctx.CtxProxyTarget)
	if !ok {
		return nil, http.StatusBadGateway, fmt.Errorf("proxy target not set: %s\n", g.contextError(ctx))
	}
	target, ok := value.(string)
	if !ok {
		return nil, http.StatusBadGateway, fmt.Errorf("proxy target not set: %s\n", g.contextError(ctx))
	}
	if target == "" {
		return nil, http.StatusBadGateway, fmt.Errorf("proxy target not set: %s\n", g.contextError(ctx))
	}
	proxyTarget, err := url.Parse(fmt.Sprintf("%s%s", target, queryParams(ctx, "&")))
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
	return proxyTarget, 0, nil
}

func (g *GinBasedGateway) contextError(ctx *gin.Context) string {
	v, ok := ctx.Get(gwctx.CtxError)
	if !ok {
		return ""
	}
	switch v.(type) {
	case error:
		return v.(error).Error()
	case string:
		return v.(string)
	default:
		return ""
	}
}

// endregion
