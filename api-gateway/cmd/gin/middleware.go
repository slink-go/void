package main

import (
	"fmt"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"
	"github.com/slink-go/api-gateway/middleware/auth"
	"github.com/slink-go/api-gateway/middleware/constants"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"net/http"
	"strings"
	"time"
)

func customLogger() gin.HandlerFunc {
	logger := logging.GetLogger("gin")
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		if latency > time.Minute {
			latency = latency.Truncate(time.Minute)
		} else {
			latency = latency.Truncate(time.Microsecond)
		}
		logger.Info("%15v %10v %7v %10v %v",
			c.ClientIP(),
			latency,
			c.Writer.Status(),
			c.Request.Method,
			c.Request.URL,
		)
	}
}

// headersCleaner - cleanup incoming headers to prevent security issues
func headersCleaner(headers ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if headers == nil || len(headers) == 0 {
			for k, _ := range ctx.Request.Header {
				if strings.HasPrefix("ctx", strings.ToLower(k)) {
					delete(ctx.Request.Header, k)
				}
			}
		} else {
			for _, h := range headers {
				delete(ctx.Request.Header, h)
			}
		}
	}
}

//func realIp() gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		ctx.Header("X-Real-Ip", ctx.ClientIP())
//	}
//}

// proxyTargetResolver - resolve request URL to target service URL
func proxyTargetResolver(reverseProxy *proxy.ReverseProxy) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := logging.GetLogger("resolver-middleware")
		logger.Trace("[resolver] handle")
		target, err := reverseProxy.ResolveTarget(ctx.Request.URL.Path)
		if err != nil {
			logger.Trace("%s", stacktrace.RootCause(err))
			switch err.(type) {
			case *resolver.ErrEmptyBaseUrl:
				ctx.Writer.WriteString(err.Error())
				ctx.Writer.WriteString("\n")
				ctx.AbortWithError(http.StatusBadRequest, err)
			case *resolver.ErrInvalidPath:
				ctx.Writer.WriteString(err.Error())
				ctx.Writer.WriteString("\n")
				ctx.AbortWithError(http.StatusBadRequest, err)
			case *registry.ErrServiceUnavailable:
				ctx.Writer.WriteString(err.Error())
				ctx.Writer.WriteString("\n")
				ctx.AbortWithStatus(http.StatusServiceUnavailable)
			}
		} else {
			logger.Trace(
				"resolved url: %s://%s%s%s -> %s",
				ctx.Request.URL.Scheme, ctx.Request.Host, ctx.Request.URL.Path, queryParams(ctx, ", "), target,
			)
			ctx.Set(constants.CtxProxyTarget, target.String())
		}
	}
}

// authResolver - resolve request authentication type/value
func authResolver(authProvider security.AuthProvider) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader(constants.HdrAuthorization)
		cookie, _ := ctx.Cookie(constants.HdrAuthToken)
		auth, err := authProvider.Get(header, cookie)
		if err == nil && auth != nil {
			switch auth.GetType() {
			case security.TypeBearer:
				fallthrough
			case security.TypeCookie:
				ctx.Set(constants.RequestContextAuth, auth)
			}
		}
	}
}

func authCache(cache auth.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cache == nil {
			return
		}
		var auth security.Auth
		if v, ok := ctx.Get(constants.RequestContextAuth); ok {
			if auth, ok = v.(security.Auth); !ok || auth == nil {
				return
			}
		}
		if auth == nil {
			return
		}
		v, ok := cache.Get(fmt.Sprintf("%v", auth.GetValue()))
		if !ok {
			return
		}
		ctx.Set(constants.RequestContextUserDetails, v)
	}
}

// authProvider - provide authorized user details for given Auth
func authProvider(userDetailsProvider security.UserDetailsProvider, cache auth.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if v, ok := ctx.Get(constants.RequestContextUserDetails); ok {
			if _, ok := v.(security.UserDetails); ok {
				return
			}
		}
		var auth security.Auth
		if v, ok := ctx.Get(constants.RequestContextAuth); ok {
			if auth, ok = v.(security.Auth); !ok || auth == nil {
				return
			}
		}
		if auth == nil {
			return
		}
		switch auth.GetType() {
		case security.TypeBearer:
			fallthrough
		case security.TypeCookie:
			token := auth.GetValue().(string)
			userDetails, err := userDetailsProvider.Get(token)
			if err != nil {
				logging.GetLogger("middleware").Warning("%s", stacktrace.RootCause(err))
			}
			if userDetails != nil {
				ctx.Set(constants.RequestContextUserDetails, userDetails)
				if cache != nil {
					cache.Set(token, userDetails)
				}
			}
		}
	}
}

// localeResolver - resolve request locale
func localeResolver() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		l := ctx.Request.URL.Query().Get("lang")
		if l != "" {
			ctx.Set(constants.RequestContextLocale, l)
			return
		}
		l = ctx.Request.URL.Query().Get("locale")
		if l != "" {
			ctx.Set(constants.RequestContextLocale, l)
			return
		}
		if ctx.GetHeader(constants.HdrAcceptLanguage) != "" {
			ctx.Set(constants.RequestContextLocale, ctx.GetHeader(constants.HdrAcceptLanguage))
		}
	}
}

// contextConfigurator - configure proxied request headers
func contextConfigurator() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		locale, ok := ctx.Get(constants.RequestContextLocale)
		if ok {
			ctx.Request.Header.Set(constants.HdrAcceptLanguage, locale.(string))
		}
		ud, ok := ctx.Get(constants.RequestContextUserDetails)
		if !ok {
			return
		}
		userDetails, ok := ud.(security.UserDetails)
		if ok {
			for k, v := range userDetails {
				if v != "" {
					ctx.Request.Header.Set(k, v)
				}
			}
		}
	}
}

// timeouter - ...
func timeouter(tm time.Duration) gin.HandlerFunc {
	logger := logging.GetLogger("timeout-middleware")
	return timeout.New(
		timeout.WithTimeout(tm),
		timeout.WithHandler(func(c *gin.Context) {
			logger.Trace("[timeout] handle")
			c.Next()
		}),
	)
}

func rateLimiter(lim rate.Limiter) gin.HandlerFunc {
	if lim == nil {
		return func(context *gin.Context) {
			context.Next()
		}
	}
	return func(ctx *gin.Context) {
		ctx.Set(constants.CtxRateLimiter, lim)
		lmtr := lim.Get(ctx.Request.URL.Path)
		mw := ginlimiter.NewMiddleware(
			lmtr,
			ginlimiter.WithKeyGetter(rateLimitKeyGetter),
			ginlimiter.WithLimitReachedHandler(func(c *gin.Context) {
				if lim.Mode() == rate.LimiterModeDenying {
					c.Writer.WriteHeader(http.StatusTooManyRequests)
					wait, err := getWait(lmtr, c)
					if err != nil {
						c.Writer.Write([]byte("Too many requests.\n"))
					} else {
						c.Writer.Write([]byte(fmt.Sprintf("Too many requests. Try again in %d seconds.\n", wait)))
					}
					c.Abort()
				} else {
					wait, err := getWait(lmtr, c)
					if err != nil {
						c.Writer.Write([]byte("Too many requests.\n"))
						c.Abort()
					}
					timer := time.NewTimer(time.Duration(wait) * time.Second)
					<-timer.C
					c.Next()
				}
			}),
		)
		mw(ctx)
	}
}

func getWait(lmtr *limiter.Limiter, c *gin.Context) (int64, error) {
	logger := logging.GetLogger("rate-limiter")
	key := rateLimitKeyGetter(c)
	ctx, err := lmtr.Peek(c, key)
	if err != nil {
		return 0, err
	}
	wait := ctx.Reset - time.Now().Unix()
	if wait == 0 {
		wait = 1
	}
	logger.Trace("key: %s, wait: %d", key, wait)
	return wait, nil
}
func rateLimitKeyGetter(ctx *gin.Context) string {
	realIp := ctx.ClientIP()
	v, ok := ctx.Get(constants.CtxRateLimiter)
	if !ok {
		return realIp
	}
	lim, ok := v.(rate.Limiter)
	if !ok {
		return realIp
	}
	return realIp + ":" + lim.KeyForPath(ctx.Request.URL.Path)
}

//func circuitBreakerMiddleware()
