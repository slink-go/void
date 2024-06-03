package main

import (
	"fmt"
	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-gonic/gin"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/cmd/common/templates"
	"github.com/slink-go/api-gateway/gateway"
	"github.com/slink-go/api-gateway/middleware/auth"
	"github.com/slink-go/api-gateway/middleware/constants"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/logging"
	"net/http"
	"net/url"
)

type GinBasedGateway struct {
	logger              logging.Logger
	authProvider        security.AuthProvider
	userDetailsProvider security.UserDetailsProvider
	authCache           auth.Cache
	reverseProxy        *proxy.ReverseProxy
	registry            registry.ServiceRegistry

	quitChn chan struct{}
}

//region - initializers

func NewGinBasedGateway() gateway.Gateway {
	gw := GinBasedGateway{
		logger: logging.GetLogger("gin-gateway"),
	}
	return &gw
}
func (g *GinBasedGateway) WithAuthProvider(ap security.AuthProvider) gateway.Gateway {
	g.authProvider = ap
	return g
}
func (g *GinBasedGateway) WithUserDetailsProvider(udp security.UserDetailsProvider) gateway.Gateway {
	g.userDetailsProvider = udp
	return g
}
func (g *GinBasedGateway) WithUserDetailsCache(cache auth.Cache) gateway.Gateway {
	g.authCache = cache
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

	if env.BoolOrDefault(env.MonitoringEnabled, false) {
		if len(addresses) > 1 && addresses[1] != "" {
			go NewService("monitor").
				//WithHandler("/monitor", monitor.New(monitor.Config{Title: "VOID API Gateway (monitoring)"})) // TODO: fiber-like monitoring
				WithGetHandlers("/", g.monitoringPage).
				WithGetHandlers("/list", g.listRemotes).
				WithStatic("/s", "./static").
				Run(addresses[1])
		} else {
			g.logger.Warning("no monitoring port set; disable monitoring")
		}
	}
	if addresses[0] != "" {
		authEnabled := env.BoolOrDefault(env.AuthEnabled, false)
		NewService("proxy").
			WithPrometheus().
			WithMiddleware(helmet.Default()). // TODO: custom helmet config
			//WithMiddleware(csrf.New()). // TODO: implement it for Gin (?)
			//WithMiddleware(timeoutMiddleware(100 * time.Millisecond)). // TODO: skip paths
			//WithMiddleware(rateLimiter()).
			WithMiddleware(proxyTargetResolver(g.reverseProxy)).
			//WithMiddleware(circuitBreaker()).
			WithMiddleware(headersCleaner()). // cleanup incoming headers
			WithOptionalMiddleware(authEnabled, authResolver(g.authProvider)).
			WithOptionalMiddleware(authEnabled, authCache(g.authCache)).
			WithOptionalMiddleware(authEnabled, authProvider(g.userDetailsProvider, g.authCache)).
			WithMiddleware(localeResolver()).
			WithMiddleware(contextConfigurator()).
			WithNoRouteHandler(g.proxyHandler).
			WithQuitChn(g.quitChn).
			Run(addresses[0])

	} else {
		panic("service port not set")
	}
}

// endregion
// region - monitoring

func (g *GinBasedGateway) monitoringPage(ctx *gin.Context) {
	ctx.Set(constants.HdrContentType, "text/html")
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
	value, ok := ctx.Get(constants.CtxProxyTarget)
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
	v, ok := ctx.Get(constants.CtxError)
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
