package main

import (
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"
	"github.com/slink-go/api-gateway/middleware/context"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"net/http"
	"time"
)

// resolve proxy target from URL Path
func proxyTargetResolverMiddleware(reverseProxy *proxy.ReverseProxy) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := logging.GetLogger("resolver-middleware")
		logger.Trace("[resolver] handle")
		target, err := reverseProxy.ResolveTarget(ctx.Request.URL.Path)
		if err != nil {
			logger.Trace("%s", stacktrace.RootCause(err))
			// - если смогли нужный найти сервис в реестре, устанавливаем соответствующий
			// заголовок; иначе просто продолжаем выполнение (пробуем локальный обработчик
			// вместо прокси)
			// - для полноты картины выставляем ошибку в контексте
			switch err.(type) {
			case *resolver.ErrEmptyBaseUrl:
				ctx.Set(context.CtxError, err.Error())
			case *resolver.ErrInvalidPath:
				ctx.Set(context.CtxError, err.Error())
			case *registry.ErrServiceUnavailable:
				ctx.Set(context.CtxError, err.Error())
			}
		} else {
			logger.Trace(
				"resolved url: %s://%s%s%s -> %s",
				"http", ctx.Request.Host, ctx.Request.URL.Path, queryParams(ctx, ", "), target,
			)
			ctx.Set(context.CtxProxyTarget, target.String())
		}
	}
}

// set context according to resolved target & other parameters
func contextMiddleware(contextProvider context.Provider) gin.HandlerFunc {

	return func(ctx *gin.Context) {

		logger := logging.GetLogger("context-middleware")
		logger.Trace("[context] handle")

		if _, ok := ctx.Get(context.CtxError); ok /*&& (v.(string) == "" || v == nil)*/ {
			// нет смысла пытаться установить контекст, если не
			// удалось зарезолвить сервис из реестра
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}

		lang := getQueryParam(ctx, "lang")

		cc := contextProvider.GetContext(
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
}

func timeoutMiddleware(tm time.Duration) gin.HandlerFunc {
	logger := logging.GetLogger("timeout-middleware")
	return timeout.New(
		timeout.WithTimeout(tm),
		timeout.WithHandler(func(c *gin.Context) {
			logger.Trace("[timeout] handle")
			c.Next()
		}),
	)
}

//func circuitBreakerMiddleware()
