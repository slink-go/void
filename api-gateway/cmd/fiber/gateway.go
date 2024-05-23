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
	"strings"
	"time"
)

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
	engine          *fiber.App
	registry        registry.ServiceRegistry
}

func (g *FiberBasedGateway) Serve(address string) {
	g.engine = fiber.New()
	g.setupMiddleware(address)
	g.setupRouteHandlers()
	if err := g.engine.Listen(address); err != nil {
		panic(err)
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

func (g *FiberBasedGateway) setupMiddleware(address string) {

	g.engine.Use(pprof.New())

	p := fiberprometheus.New(fmt.Sprintf("fiber-api-gateway:%s", address))
	p.RegisterAt(g.engine, "/prometheus")
	//prometheus.SetSkipPaths([]string{"/ping"})
	g.engine.Use(p.Middleware)

	// logger | должен быть первым, чтобы фиксировать отлупы от другого middleware и latency запросов
	// кстати о latency - нужна metricMiddleware
	g.engine.Use(logger.New())
	//g.engine.Use(g.customLoggingMiddleware)

	// rate limiter
	if g.limiter != nil {
		g.engine.Use(g.limiter)
	}

	// helmet (security)
	g.engine.Use(helmet.New())

	// csrf
	csrfConfig := csrf.Config{
		KeyLookup:      "header:X-Csrf-Token", // string in the form of '<source>:<key>' that is used to extract token from the request
		CookieName:     "my_csrf_",            // name of the session cookie
		CookieSameSite: "Strict",              // indicates if CSRF cookie is requested by SameSite
		Expiration:     3 * time.Hour,         // expiration is the duration before CSRF token will expire
		KeyGenerator:   utils.UUID,            // creates a new CSRF token
	}
	g.engine.Use(csrf.New(csrfConfig))

	// target resolver
	g.engine.Use(g.proxyTargetResolver)

	// context provider
	g.engine.Use(g.contextMiddleware)

}
func (g *FiberBasedGateway) setupRouteHandlers() {
	g.engine.Get("/x", g.monitoringPage)
	g.engine.Static("/x/s", "./static")
	g.engine.Get("/x/list", g.listRemotes)
	g.engine.Get("/x/monitor", monitor.New(monitor.Config{Title: "VOID API Gateway Monitor"}))
	g.engine.Get("*", g.proxyHandler)
	g.engine.Post("*", g.proxyHandler)
	g.engine.Put("*", g.proxyHandler)
	g.engine.Delete("*", g.proxyHandler)
	g.engine.Head("*", g.proxyHandler)
	g.engine.Options("*", g.proxyHandler)
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
func (g *FiberBasedGateway) customLoggingMiddleware(ctx *fiber.Ctx) error {
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
