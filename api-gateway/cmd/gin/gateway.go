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
	limiter             rate.Limiter
	quitChn             chan struct{}
}

// region - options

type Option interface {
	apply(*GinBasedGateway)
}

// region -> auth provider

type authProviderOption struct {
	value security.AuthProvider
}

func (o *authProviderOption) apply(g *GinBasedGateway) {
	if o.value != nil {
		g.authProvider = o.value
	}
}
func WithAuthProvider(value security.AuthProvider) Option {
	return &authProviderOption{value}
}

// endregion
// region -> user details provider

type userDetailsProviderOption struct {
	value security.UserDetailsProvider
}

func (o *userDetailsProviderOption) apply(g *GinBasedGateway) {
	if o.value != nil {
		g.userDetailsProvider = o.value
	}
}
func WithUserDetailsProvider(value security.UserDetailsProvider) Option {
	return &userDetailsProviderOption{value}
}

// endregion
// region -> user details cache

type userDetailsCacheOption struct {
	value auth.Cache
}

func (o *userDetailsCacheOption) apply(g *GinBasedGateway) {
	if o.value != nil {
		g.authCache = o.value
	}
}
func WithUserDetailsCache(value auth.Cache) Option {
	return &userDetailsCacheOption{value}
}

// endregion
// region -> reverse proxy

type reverseProxyOption struct {
	value *proxy.ReverseProxy
}

func (o *reverseProxyOption) apply(g *GinBasedGateway) {
	if o.value != nil {
		g.reverseProxy = o.value
	}
}
func WithReverseProxy(value *proxy.ReverseProxy) Option {
	return &reverseProxyOption{value}
}

// endregion
// region -> rate limiter

type rateLimiterOption struct {
	value rate.Limiter
}

func (o *rateLimiterOption) apply(g *GinBasedGateway) {
	if o.value != nil {
		g.limiter = o.value
	}
}
func WithRateLimiter(value rate.Limiter) Option {
	return &rateLimiterOption{value}
}

// endregion
// region -> registry

type registryOption struct {
	value registry.ServiceRegistry
}

func (o *registryOption) apply(g *GinBasedGateway) {
	if o.value != nil {
		g.registry = o.value
	}
}
func WithRegistry(value registry.ServiceRegistry) Option {
	return &registryOption{value}
}

// endregion
// region -> quit chn

type quitChnOption struct {
	value chan struct{}
}

func (o *quitChnOption) apply(g *GinBasedGateway) {
	if o.value != nil {
		g.quitChn = o.value
	}
}
func WithQuitChn(value chan struct{}) Option {
	return &quitChnOption{value}
}

// endregion

// endregion
//region - initializers

func NewGinBasedGateway(options ...Option) gateway.Gateway {
	gw := GinBasedGateway{
		logger: logging.GetLogger("gin-gateway"),
	}
	for _, option := range options {
		if option != nil {
			option.apply(&gw)
		}
	}
	return &gw
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
			WithMiddleware(gin.Recovery()).
			WithMiddleware(customLogger()).
			WithMiddleware(headersCleaner()).
			WithMiddleware(rateLimiter(g.limiter)).
			WithMiddleware(helmet.Default()). // TODO: custom helmet config
			//WithMiddleware(csrf.New()). // TODO: implement it for Gin (?)
			//WithMiddleware(timeouter(100 * time.Millisecond)). // TODO: skip paths
			WithMiddleware(proxyTargetResolver(g.reverseProxy)).
			//WithMiddleware(circuitBreaker()).
			WithOptionalMiddleware(authEnabled, authResolver(g.authProvider)).
			WithOptionalMiddleware(authEnabled, authCache(g.authCache)).
			WithOptionalMiddleware(authEnabled, authProvider(g.userDetailsProvider, g.authCache)).
			WithMiddleware(localeResolver()).
			WithMiddleware(contextConfigurator()).
			WithNoRouteHandlers(g.proxyHandler).
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

	//ctx.Set("X-Forwarded-For", ctx.RemoteIP())
	//ctx.Set("X-Real-Ip", ctx.ClientIP())

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
