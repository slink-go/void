package main

import (
	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"
	"github.com/slink-go/api-gateway/gateway"
	"github.com/slink-go/api-gateway/middleware/context"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"go.uber.org/ratelimit"
	"net/http"
	"strings"
)

func NewGinBasedGateway() gateway.Gateway {
	gateway := GinBasedGateway{
		logger:          logging.GetLogger("gin-gateway"),
		contextProvider: context.CreateContextProvider(),
	}
	return &gateway
}

type GinBasedGateway struct {
	logger          logging.Logger
	contextProvider context.Provider
	reverseProxy    *proxy.ReverseProxy
	limiter         ratelimit.Limiter
	engine          *gin.Engine
}

func (g *GinBasedGateway) Serve(address string) {
	g.setupRoutes()
	g.engine.Run(address)
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
func (g *GinBasedGateway) WithRateLimiter(limiter rate.Limiter) gateway.Gateway {
	g.limiter = ratelimit.New(limiter.GetLimit())
	return g
}

func (g *GinBasedGateway) setupRoutes() {
	g.engine = gin.New()
	g.setupRateLimiter()
	g.engine.GET("*path", g.contextSet, g.proxyHandler)
	g.engine.POST("*path", g.contextSet, g.proxyHandler)
	g.engine.PUT("*path", g.contextSet, g.proxyHandler)
	g.engine.DELETE("*path", g.contextSet, g.proxyHandler)
	g.engine.HEAD("*path", g.contextSet, g.proxyHandler)
	g.engine.PATCH("*path", g.contextSet, g.proxyHandler)
	g.engine.OPTIONS("*path", g.contextSet, g.proxyHandler)
}

func (g *GinBasedGateway) setupRateLimiter() {
	if g.limiter != nil {
		g.engine.Use(g.leakBucket())
	}
}

func (g *GinBasedGateway) leakBucket() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		g.limiter.Take()
	}
}

func (g *GinBasedGateway) contextSet(ctx *gin.Context) {
	lang := ""
	if v, ok := ctx.Request.URL.Query()["lang"]; ok {
		lang = v[0]
	}
	cc := g.contextProvider.GetContext(
		context.NewAuthContextOption(ctx.GetHeader("Authorization")),
		context.NewLocalizationOption(ctx.GetHeader("Accept-Language")),
		context.NewLangParamOption(lang),
	)
	for k, v := range cc {
		if len(v) > 0 {
			ctx.Request.Header.Set(k, v[0])
		}
		if len(v) > 1 {
			for _, h := range v[1:] {
				ctx.Request.Header.Set(k, h)
			}
		}
	}
	if v, ok := cc[context.CtxLocale]; ok && len(v) > 0 {
		ctx.Request.Header.Set("Accept-Language", v[0])
	}

}
func (g *GinBasedGateway) proxyHandler(ctx *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			g.logger.Warning("panic: %v", err)
		}
	}()

	target, err := g.reverseProxy.ResolveTarget(ctx.Request.URL.Path)
	if err != nil {
		g.logger.Warning("%s", stacktrace.RootCause(err))
		switch err.(type) {
		case *resolver.ErrEmptyBaseUrl:
			ctx.AbortWithStatus(http.StatusBadGateway)
		case *resolver.ErrInvalidPath:
			ctx.AbortWithStatus(http.StatusBadRequest)
		case *registry.ErrServiceUnavailable:
			ctx.AbortWithStatus(http.StatusServiceUnavailable)
		}
		return
	}
	g.logger.Trace("resolved url: %s://%s%s%s -> %s", "http", ctx.Request.Host, ctx.Request.URL.Path, g.queryParams(ctx), target)

	// TODO: implement it
	//headers, err := preprocessRequest(ctx)
	//if err != nil {
	//	ctx.AbortWithStatus(http.StatusUnauthorized)
	//}
	g.reverseProxy.Proxy(target).ServeHTTP(ctx.Writer, ctx.Request)
}
func (g *GinBasedGateway) queryParams(ctx *gin.Context) string {
	result := ""
	params := ctx.Request.URL.Query()
	if len(params) > 0 {
		for k, p := range params {
			for _, v := range p {
				result = result + k
				result = result + "="
				result = result + v
				result = result + ", "
			}
		}
		result = "?" + strings.TrimSuffix(result, ", ")
	}
	return result
}
