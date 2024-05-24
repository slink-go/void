package main

import (
	"encoding/json"
	"fmt"
	"github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/csrf"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	p "github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/gofiber/fiber/v2/utils"
	"github.com/palantir/stacktrace"
	"github.com/slink-go/api-gateway/cmd/fiber/templates"
	"github.com/slink-go/api-gateway/gateway"
	"github.com/slink-go/api-gateway/middleware/context"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// see also:
// https://stackoverflow.com/questions/76501736/how-to-prevent-fiber-from-auto-registerting-head-routes
// https://github.com/renanbastos93/fastpath?tab=readme-ov-file
// https://medium.com/@bijit211987/everything-you-need-to-know-about-rate-limiting-for-apis-f236d2adcfff
// https://docs.spring.io/spring-cloud/docs/current/reference/html/configprops.html
// PROFILING https://habr.com/ru/companies/badoo/articles/324682/
// !!! https://dev.to/koddr/go-fiber-by-examples-working-with-middlewares-and-boilerplates-3p0m#helmet-middleware
// !!! Create Go App https://github.com/create-go-app

func NewFiberBasedGateway() gateway.Gateway {
	gw := FiberBasedGateway{
		logger:          logging.GetLogger("fiber-gateway"),
		contextProvider: context.CreateContextProvider(),
	}
	return &gw
}

type FiberBasedGateway struct {
	logger          logging.Logger
	contextProvider context.Provider
	reverseProxy    *proxy.ReverseProxy
	limiter         fiber.Handler
	registry        registry.ServiceRegistry
	service         *fiber.App
	monitoring      *fiber.App
	quitChn         chan<- struct{}
}

// region - initializers

func (g *FiberBasedGateway) Serve(addresses ...string) {
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

func (g *FiberBasedGateway) WithAuthProvider(ap security.AuthProvider) gateway.Gateway {
	g.contextProvider.WithAuthProvider(ap)
	return g
}
func (g *FiberBasedGateway) WithUserDetailsProvider(udp security.UserDetailsProvider) gateway.Gateway {
	g.contextProvider.WithUserDetailsProvider(udp)
	return g
}
func (g *FiberBasedGateway) WithReverseProxy(reverseProxy *proxy.ReverseProxy) gateway.Gateway {
	g.reverseProxy = reverseProxy
	return g
}
func (g *FiberBasedGateway) WithRateLimiter(limit rate.Limiter) gateway.Gateway {
	limiterConfig := limiter.Config{
		Max:                    limit.GetLimit(),
		Expiration:             time.Minute,
		SkipFailedRequests:     false,
		SkipSuccessfulRequests: false,
		LimiterMiddleware:      limiter.FixedWindow{},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(http.StatusTooManyRequests).Send([]byte(fmt.Sprintf("Too Many Requests\n")))
		},
		KeyGenerator: func(c *fiber.Ctx) string {
			proxyIP := c.Get("X-Forwarded-For")
			if proxyIP != "" {
				return proxyIP
			} else {
				return c.IP()
			}
		},
	}
	g.limiter = limiter.New(limiterConfig)
	return g
}
func (g *FiberBasedGateway) WithRegistry(registry registry.ServiceRegistry) gateway.Gateway {
	g.registry = registry
	return g
}
func (g *FiberBasedGateway) WithQuitChn(chn chan struct{}) gateway.Gateway {
	g.quitChn = chn
	return g
}

// endregion
// region - monitoring

func (g *FiberBasedGateway) startMonitoring(address string) {

	g.monitoring = fiber.New()
	g.setupMonitoringMiddleware(address)
	g.setupMonitoringRouteHandlers()
	g.logger.Info("start monitoring service on %s", address)

	go g.handleBreak("monitoring", g.monitoring)

	if err := g.monitoring.Listen(address); err != nil {
		panic(err)
	}

}

func (g *FiberBasedGateway) setupMonitoringMiddleware(address string) {
	p := fiberprometheus.New(fmt.Sprintf("fiber-api-gateway-monitor:%s", address))
	p.RegisterAt(g.monitoring, "/prometheus")
	g.monitoring.Use(p.Middleware)
	g.monitoring.Use(logger.New())
}

func (g *FiberBasedGateway) setupMonitoringRouteHandlers() {
	g.monitoring.Get("/", g.monitoringPage)
	g.monitoring.Static("/s", "./static")
	g.monitoring.Get("/list", g.listRemotes)
	g.monitoring.Get("/monitor", monitor.New(monitor.Config{Title: "VOID API Gateway (monitoring)"}))
}
func (g *FiberBasedGateway) monitoringPage(c *fiber.Ctx) error {
	c.Set("Content-Type", "text/html")
	t := templates.ServicesPage(templates.Cards(g.registry.List()))
	err := t.Render(c.Context(), c.Response().BodyWriter())
	return err
}
func (g *FiberBasedGateway) listRemotes(c *fiber.Ctx) error {
	if g.registry != nil {
		data := g.registry.List()
		if data == nil || len(data) == 0 {
			c.Status(fiber.StatusNoContent)
		} else {
			buff, err := json.Marshal(data)
			if err != nil {
				c.Status(fiber.StatusInternalServerError).Write([]byte(err.Error()))
			} else {
				c.Write(buff)
			}
		}
	} else {
		c.Status(fiber.StatusNoContent)
	}
	return nil
}

// endregion
// region - service

func (g *FiberBasedGateway) startProxyService(address string) {

	g.service = fiber.New()
	g.setupMiddleware(address)
	g.setupRouteHandlers()
	g.logger.Info("start proxy service on %s", address)

	go g.handleBreak("proxy", g.service)
	if err := g.service.Listen(address); err != nil {
		panic(err)
	}

}

func (g *FiberBasedGateway) setupMiddleware(address string) {

	g.service.Use(pprof.New())

	p := fiberprometheus.New(fmt.Sprintf("fiber-api-gateway:%s", address))
	p.RegisterAt(g.service, "/prometheus")
	//prometheus.SetSkipPaths([]string{"/ping"})
	g.service.Use(p.Middleware)

	// logger | должен быть первым, чтобы фиксировать отлупы от другого middleware и latency запросов
	// кстати о latency - нужна metricMiddleware
	g.service.Use(logger.New())
	//g.engine.Use(g.customLoggingMiddleware)

	// rate limiter
	if g.limiter != nil {
		g.service.Use(g.limiter)
	}

	// helmet (security)
	g.service.Use(helmet.New())

	// csrf
	csrfConfig := csrf.Config{
		KeyLookup:      "header:X-Csrf-Token", // string in the form of '<source>:<key>' that is used to extract token from the request
		CookieName:     "my_csrf_",            // name of the session cookie
		CookieSameSite: "Strict",              // indicates if CSRF cookie is requested by SameSite
		Expiration:     3 * time.Hour,         // expiration is the duration before CSRF token will expire
		KeyGenerator:   utils.UUID,            // creates a new CSRF token
	}
	g.service.Use(csrf.New(csrfConfig))

	// target resolver
	g.service.Use(g.proxyTargetResolver)

	// context provider
	g.service.Use(g.contextMiddleware)

}
func (g *FiberBasedGateway) customLoggingMiddleware(ctx *fiber.Ctx) error {
	return ctx.Next()
}
func (g *FiberBasedGateway) proxyTargetResolver(ctx *fiber.Ctx) error {
	target, err := g.reverseProxy.ResolveTarget(ctx.Path())
	if err != nil {
		g.logger.Trace("%s", stacktrace.RootCause(err))
		// - если смогли нужный найти сервис в реестре, устанавливаем соответствующий
		// заголовок; иначе просто продолжаем выполнение (пробуем локальный обработчик
		// вместо прокси)
		// - для полноты картины выставляем ошибку в контексте
		switch err.(type) {
		case *resolver.ErrEmptyBaseUrl:
			ctx.Request().Header.Set(context.CtxError, err.Error())
		case *resolver.ErrInvalidPath:
			ctx.Request().Header.Set(context.CtxError, err.Error())
		case *registry.ErrServiceUnavailable:
			ctx.Request().Header.Set(context.CtxError, err.Error())
		}
	} else {
		g.logger.Trace("resolved url: %s%s -> %s", ctx.BaseURL(), ctx.OriginalURL(), target)
		ctx.Request().Header.Set(context.CtxProxyTarget, target.String())
	}
	return ctx.Next()
}
func (g *FiberBasedGateway) contextMiddleware(ctx *fiber.Ctx) error {
	if getHeader(ctx, context.CtxError) != "" {
		// нет смысла пытаться установить контекст, если не
		// удалось зарезолвить сервис из реестра
		return ctx.Next()
	}
	lang := ctx.Query("lang", "")
	cc := g.contextProvider.GetContext(
		context.NewAuthContextOption(getHeader(ctx, "Authorization")),
		context.NewLocalizationOption(getHeader(ctx, "Accept-Language")),
		context.NewLangParamOption(lang),
	)
	for k, v := range cc {
		if len(v) > 0 {
			ctx.Request().Header.Set(k, v[0])
		}
		if len(v) > 1 {
			for _, h := range v[1:] {
				ctx.Request().Header.Set(k, h)
			}
		}
	}
	if v, ok := cc[context.CtxLocale]; ok && len(v) > 0 {
		ctx.Request().Header.Set("Accept-Language", v[0])
	}
	return ctx.Next()
}

func (g *FiberBasedGateway) setupRouteHandlers() {
	g.service.Get("/monitor", monitor.New(monitor.Config{Title: "VOID API Gateway (service)"}))
	g.service.Get("*", g.proxyHandler)
	g.service.Post("*", g.proxyHandler)
	g.service.Put("*", g.proxyHandler)
	g.service.Delete("*", g.proxyHandler)
	g.service.Head("*", g.proxyHandler)
	g.service.Options("*", g.proxyHandler)
}
func (g *FiberBasedGateway) proxyHandler(ctx *fiber.Ctx) error {

	defer func() {
		// а как вернуть RequestScoped-ошибку?
		if err := recover(); err != nil {
			g.logger.Warning("panic: %v", err)
		}
	}()

	target := getHeader(ctx, context.CtxProxyTarget)
	if target == "" {
		ctx.Status(http.StatusBadGateway).Write([]byte(
			fmt.Sprintf("proxy target not set: %s\n", ctx.Get(context.CtxError))),
		)
		return nil
	}

	proxyTarget := fmt.Sprintf("%s%s", target, g.getQueryParams(ctx))
	g.logger.Trace("proxying %s", proxyTarget)

	// TODO: сделать обработку ответов !!! (а то сильно прозрачно получается)
	return p.Do(ctx, proxyTarget)
}
func (g *FiberBasedGateway) getQueryParams(ctx *fiber.Ctx) string {
	result := ""
	for p, v := range ctx.Queries() {
		result = result + p
		result = result + "="
		result = result + v
		result = result + "&"
	}
	return "?" + strings.TrimSuffix(result, "&")
}
func getHeader(ctx *fiber.Ctx, key string) string {
	if v, ok := ctx.GetReqHeaders()[key]; ok {
		return v[0]
	}
	return ""
}

// endregion
// region - common

func (g *FiberBasedGateway) handleBreak(service string, app *fiber.App) {
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
			app.Shutdown()
			close(sigChn)
			if g.quitChn != nil {
				g.quitChn <- struct{}{}
			}
			return
		}
	}
}

// endregion
