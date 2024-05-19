package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	p "github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/palantir/stacktrace"
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"net/http"
	"strings"
	"time"
)

// https://docs.gofiber.io/category/-middleware/

var logger logging.Logger
var reverseProxy *proxy.ReverseProxy

func main() {
	common.LoadEnv()

	reverseProxy = proxy.CreateReverseProxy().
		WithServiceResolver(common.ServiceResolver()).
		WithPathProcessor(resolver.NewPathProcessor())

	router := fiber.New()
	router.Use(limiter.New(limiter.Config{
		Max:                    5,
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
	}))
	router.Get("*", proxyHandler)
	router.Post("*", proxyHandler)
	router.Listen(":3004")
}

func proxyHandler(ctx *fiber.Ctx) error {

	defer func() {
		if err := recover(); err != nil {
			logger.Warning("panic: %v", err)
		}
	}()

	target, err := reverseProxy.ResolveTarget(ctx.Path())
	if err != nil {
		logger.Warning("%s", stacktrace.RootCause(err))
		switch err.(type) {
		case *resolver.ErrEmptyBaseUrl:
			ctx.Status(http.StatusBadGateway)
		case *resolver.ErrInvalidPath:
			ctx.Status(http.StatusBadRequest)
		case *resolver.ErrServiceUnavailable:
			ctx.Status(http.StatusServiceUnavailable)
		}
		return err
	}
	logger.Info("resolved url: %s%s -> %s", ctx.BaseURL(), ctx.OriginalURL(), target)
	ctx.Queries()

	// TODO: implement it
	//headers, err := preprocessRequest(ctx)
	//if err != nil {
	//	ctx.AbortWithStatus(http.StatusUnauthorized)
	//}
	ctx.Request().Header.Set("Header", "STUB")

	// TODO: сделать обработку ответов !!! (а то сильно прозрачно получается)
	return p.Do(ctx, fmt.Sprintf("%s?%s", target.String(), queryParams(ctx)))
}
func queryParams(c *fiber.Ctx) string {
	result := ""
	for p, v := range c.Queries() {
		result = result + p
		result = result + "="
		result = result + v
		result = result + ", "
	}
	return strings.TrimSuffix(result, ", ")
}
