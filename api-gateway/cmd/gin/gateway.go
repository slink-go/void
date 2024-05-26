package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slink-go/api-gateway/cmd/common/templates"
	"github.com/slink-go/api-gateway/gateway"
	gwctx "github.com/slink-go/api-gateway/middleware/context"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type GinBasedGateway struct {
	logger          logging.Logger
	contextProvider gwctx.Provider
	reverseProxy    *proxy.ReverseProxy
	registry        registry.ServiceRegistry
	proxy           *gin.Engine
	monitoring      *gin.Engine
	quitChn         chan<- struct{}
	//limiter         fiber.Handler
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
	defer func() {
		if g.quitChn != nil {
			g.quitChn <- struct{}{}
		}
		if err := recover(); err != nil {
			g.logger.Warning("server error: ")
		}
	}()
	if addresses == nil || len(addresses) == 0 {
		panic("service address(es) not set")
	}
	if len(addresses) > 1 && addresses[1] != "" {
		go g.startMonitoring(addresses[1])
	}
	if addresses[0] != "" {
		g.startProxyService(addresses[0])
	} else {
		panic("service port not set")
	}
}

// endregion
//region - monitoring

func (g *GinBasedGateway) startMonitoring(address string) {

	g.monitoring = gin.Default()
	g.setupMonitoringMiddleware()
	g.setupMonitoringRouteHandlers()
	server := &http.Server{
		Addr:    address,
		Handler: g.monitoring,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.logger.Panic("[monitoring]listen: %s\n", err)
		}
	}()
	g.logger.Info("start monitoring service on %s", address)
	g.handleBreak("monitoring", server)

}

func (g *GinBasedGateway) setupMonitoringMiddleware() {
	g.enablePrometheus(g.monitoring)
}
func (g *GinBasedGateway) setupMonitoringRouteHandlers() {
	g.monitoring.GET("/", g.monitoringPage)
	g.monitoring.Static("/s", "./static")
	g.monitoring.GET("/list", g.listRemotes)
	// TODO: fiber-like monitoring
	//g.monitoring.GET("/monitor", monitor.New(monitor.Config{Title: "VOID API Gateway (monitoring)"}))
}
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
//region - proxy

func (g *GinBasedGateway) startProxyService(address string) {

	g.proxy = gin.Default()
	g.setupProxyMiddleware()
	g.setupProxyRouteHandlers()
	server := &http.Server{
		Addr:    address,
		Handler: g.monitoring,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.logger.Panic("[proxy] listen: %s\n", err)
		}
	}()
	g.logger.Info("start proxy service on %s", address)
	g.handleBreak("proxy", server)

}

func (g *GinBasedGateway) setupProxyMiddleware() {

	g.enablePrometheus(g.proxy)

	//g.proxy.Use(pprof.New())

	//p := fiberprometheus.New(fmt.Sprintf("fiber-api-gateway:%s", address))
	//p.RegisterAt(g.service, "/prometheus")
	//prometheus.SetSkipPaths([]string{"/ping"})
	//g.proxy.Use(p.Middleware)

	// logger | должен быть первым, чтобы фиксировать отлупы от другого middleware и latency запросов
	// кстати о latency - нужна metricMiddleware
	//g.service.Use(logger.New())
	//g.engine.Use(g.customLoggingMiddleware)

	// rate limiter
	//if g.limiter != nil {
	//	g.service.Use(g.limiter)
	//}

	// helmet (security)
	//g.proxy.Use(helmet.New())

	// csrf
	//csrfConfig := csrf.Config{
	//	KeyLookup:      "header:X-Csrf-Token", // string in the form of '<source>:<key>' that is used to extract token from the request
	//	CookieName:     "my_csrf_",            // name of the session cookie
	//	CookieSameSite: "Strict",              // indicates if CSRF cookie is requested by SameSite
	//	Expiration:     3 * time.Hour,         // expiration is the duration before CSRF token will expire
	//	KeyGenerator:   utils.UUID,            // creates a new CSRF token
	//}
	//g.proxy.Use(csrf.New(csrfConfig))

	// target resolver
	g.proxy.Use(g.proxyTargetResolver)

	// context provider
	g.proxy.Use(g.contextMiddleware)

}
func (g *GinBasedGateway) customLoggingMiddleware(ctx *gin.Context) {
}
func (g *GinBasedGateway) proxyTargetResolver(ctx *gin.Context) {
	target, err := g.reverseProxy.ResolveTarget(ctx.Request.URL.Path)
	if err != nil {
		g.logger.Trace("%s", stacktrace.RootCause(err))
		// - если смогли нужный найти сервис в реестре, устанавливаем соответствующий
		// заголовок; иначе просто продолжаем выполнение (пробуем локальный обработчик
		// вместо прокси)
		// - для полноты картины выставляем ошибку в контексте
		switch err.(type) {
		case *resolver.ErrEmptyBaseUrl:
			ctx.Header(gwctx.CtxError, err.Error())
		case *resolver.ErrInvalidPath:
			ctx.Header(gwctx.CtxError, err.Error())
		case *registry.ErrServiceUnavailable:
			ctx.Header(gwctx.CtxError, err.Error())
		}
	} else {
		g.logger.Trace(
			"resolved url: %s://%s%s%s -> %s",
			"http", ctx.Request.Host, ctx.Request.URL.Path, queryParams(ctx, ", "), target,
		)
		ctx.Header(gwctx.CtxProxyTarget, target.String())
	}
}
func (g *GinBasedGateway) contextMiddleware(ctx *gin.Context) {

	if ctx.GetHeader(gwctx.CtxError) != "" {
		// нет смысла пытаться установить контекст, если не
		// удалось зарезолвить сервис из реестра
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}

	lang := getQueryParam(ctx, "lang")

	cc := g.contextProvider.GetContext(
		gwctx.NewAuthContextOption(ctx.GetHeader("Authorization")),
		gwctx.NewLocalizationOption(ctx.GetHeader("Accept-Language")),
		gwctx.NewLangParamOption(lang),
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
	if v, ok := cc[gwctx.CtxLocale]; ok && len(v) > 0 {
		ctx.Request.Header.Set("Accept-Language", v[0])
	}

}

func (g *GinBasedGateway) setupProxyRouteHandlers() {
	//g.proxy.GET("/monitor", monitor.New(monitor.Config{Title: "VOID API Gateway (service)"}))
	g.proxy.GET("*path", g.proxyHandler)
	//g.proxy.POST("*path", g.proxyHandler)
	//g.proxy.PUT("*path", g.proxyHandler)
	//g.proxy.DELETE("*path", g.proxyHandler)
	//g.proxy.HEAD("*path", g.proxyHandler)
	//g.proxy.OPTIONS("*path", g.proxyHandler)
}
func (g *GinBasedGateway) proxyHandler(ctx *gin.Context) {

	//defer func() {
	//	// а как вернуть RequestScoped-ошибку?
	//	if err := recover(); err != nil {
	//		g.logger.Warning("panic: %v", err)
	//	}
	//}()

	target := ctx.GetHeader(gwctx.CtxProxyTarget) //getHeader(ctx, )
	if target == "" {
		ctx.AbortWithError(http.StatusBadGateway, fmt.Errorf("proxy target not set: %s\n", ctx.GetHeader(gwctx.CtxError)))
		return
	}

	proxyTarget, err := url.Parse(fmt.Sprintf("%s%s", target, queryParams(ctx, "&")))
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
	g.logger.Trace("proxying %s", proxyTarget)

	//fiber.AcquireAgent().Request().
	//ctx.Set("X-Forwarded-For", ctx.Context().RemoteAddr().String())
	//ctx.Set("X-Real-Ip", ctx.Context().RemoteAddr().String())
	//g.logger.Warning("remote IP: %s", ctx.Context().RemoteAddr())

	g.reverseProxy.Proxy(proxyTarget)

}

//func (g *GinBasedGateway) getQueryParams(ctx *fiber.Ctx) string {
//	result := ""
//	for p, v := range ctx.Queries() {
//		result = result + p
//		result = result + "="
//		result = result + v
//		result = result + "&"
//	}
//	if result != "" {
//		return "?" + strings.TrimSuffix(result, "&")
//	} else {
//		return ""
//	}
//}
//func getHeader(ctx *fiber.Ctx, key string) string {
//	if v, ok := ctx.GetReqHeaders()[key]; ok {
//		return v[0]
//	}
//	return ""
//}

// endregion
// region - common

func (g *GinBasedGateway) enablePrometheus(engine *gin.Engine) {
	engine.GET("/prometheus", gin.WrapH(promhttp.Handler()))
	//p.RegisterAt(g.monitoring, "/prometheus")
	//g.monitoring.Use(p.Middleware)
	engine.Use(gin.Logger())
}

//p := fiberprometheus.New(fmt.Sprintf("fiber-api-gateway-monitor:%s", address))

func (g *GinBasedGateway) handleBreak(service string, server *http.Server) {
	sigChn := make(chan os.Signal)
	signal.Notify(sigChn, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		switch <-sigChn {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGTERM:
			g.logger.Info("shutdown %s service", service)
			close(sigChn)
			shutdown(server, g.logger)
			if g.quitChn != nil {
				g.quitChn <- struct{}{}
			}
			return
		}
	}
}

func shutdown(server *http.Server, logger logging.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Panic("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}

func getQueryParam(ctx *gin.Context, key string) string {
	v, ok := ctx.Request.URL.Query()[key]
	if !ok {
		return ""
	}
	return v[0]
}
func getQueryParams(ctx *gin.Context, key string) []string {
	v, ok := ctx.Request.URL.Query()[key]
	if !ok {
		return []string{}
	}
	return v
}
func queryParams(ctx *gin.Context, joiner string) string {
	var result []string
	params := ctx.Request.URL.Query()
	if len(params) > 0 {
		for k, p := range params {
			for _, v := range p {
				result = append(result, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}
	if len(result) > 0 {
		return ""
	}
	return fmt.Sprintf("?%s", strings.Join(result, joiner))
}

// endregion
