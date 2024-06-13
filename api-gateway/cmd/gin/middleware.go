package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/palantir/stacktrace"
	"github.com/slink-go/api-gateway/middleware/auth"
	"github.com/slink-go/api-gateway/middleware/constants"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/gin-timeout"
	"github.com/slink-go/logging"
	"github.com/slink-go/util/matcher"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"net/http"
	"strings"
	"time"
)

var authSkipMatcher matcher.PatternMatcher
var timeoutSkipMatcher matcher.PatternMatcher

// region - logger

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

// endregion
// region - headersCleaner - cleanup incoming headers to prevent security issues

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

// endregion
// region - proxy target resolver - resolve request URL to target service URL

func proxyTargetResolver(reverseProxy *proxy.ReverseProxy) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		logger := logging.GetLogger("resolver-middleware")
		logger.Trace("[resolver] handle")
		target, err := reverseProxy.ResolveTarget(ctx.Request.URL.Path)
		if err != nil {
			logger.Trace("%s", stacktrace.RootCause(err))
			switch err.(type) {
			case *resolver.ErrEmptyBaseUrl:
				_, _ = ctx.Writer.WriteString(err.Error())
				_, _ = ctx.Writer.WriteString("\n")
				_ = ctx.AbortWithError(http.StatusBadRequest, err)
			case *resolver.ErrInvalidPath:
				_, _ = ctx.Writer.WriteString(err.Error())
				_, _ = ctx.Writer.WriteString("\n")
				_ = ctx.AbortWithError(http.StatusBadRequest, err)
			case *registry.ErrServiceUnavailable:
				_, _ = ctx.Writer.WriteString(err.Error())
				_, _ = ctx.Writer.WriteString("\n")
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

// endregion
// region - authentication

func authResolver(authProvider security.AuthProvider) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if authSkipMatcher != nil && authSkipMatcher.Matches(ctx.Request.URL.Path) {
			return
		}
		header := ctx.GetHeader(constants.HdrAuthorization)
		cookie, _ := ctx.Cookie(constants.HdrAuthToken)
		authentication, err := authProvider.Get(header, cookie)
		if err == nil && authentication != nil && authentication.GetType() != security.TypeNone {
			switch authentication.GetType() {
			case security.TypeBearer:
				fallthrough
			case security.TypeCookie:
				ctx.Set(constants.RequestContextAuth, authentication)
			default:
			}
		} else {
			ctx.Writer.Write([]byte(http.StatusText(http.StatusUnauthorized)))
			ctx.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}
func authCache(cache auth.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if cache == nil {
			return
		}
		var authentication security.Auth
		if v, ok := ctx.Get(constants.RequestContextAuth); ok {
			if authentication, ok = v.(security.Auth); !ok || authentication == nil {
				return
			}
		}
		if authentication == nil {
			return
		}
		v, ok := cache.Get(fmt.Sprintf("%v", authentication.GetValue()))
		if !ok {
			return
		}
		ctx.Set(constants.RequestContextUserDetails, v)
	}
}
func authProvider(userDetailsProvider security.UserDetailsProvider, cache auth.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if v, ok := ctx.Get(constants.RequestContextUserDetails); ok {
			if _, ok := v.(security.UserDetails); ok {
				return
			}
		}
		var authentication security.Auth
		if v, ok := ctx.Get(constants.RequestContextAuth); ok {
			if authentication, ok = v.(security.Auth); !ok || authentication == nil {
				return
			}
		}
		if authentication == nil {
			return
		}
		switch authentication.GetType() {
		case security.TypeBearer:
			fallthrough
		case security.TypeCookie:
			token := authentication.GetValue().(string)
			userDetails, err := userDetailsProvider.Get(token)
			if err != nil {
				logging.GetLogger("middleware").Warning("%s", stacktrace.RootCause(err))
			}
			if userDetails != nil {
				ctx.Set(constants.RequestContextUserDetails, userDetails)
				if cache != nil {
					cache.Set(token, userDetails)
				}
			} else {
				ctx.Writer.Write([]byte(http.StatusText(http.StatusForbidden)))
				ctx.AbortWithStatus(http.StatusForbidden)
			}
		default:
		}
	}
}

// endregion
// region - locale resolver

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

// endregion
// region - context configurator

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

// endregion
// region - timeouter - ...

func timeouter(tm time.Duration, skipPatterns matcher.PatternMatcher) gin.HandlerFunc {
	//logger := logging.GetLogger("timeout-middleware")
	return timeout.New(
		timeout.WithTimeout(tm),
		timeout.WithSkip(skipPatterns),
		timeout.WithHandler(func(context *gin.Context) {
			context.Next()
		}),
		timeout.WithResponse(func(context *gin.Context) {
			context.String(http.StatusRequestTimeout, http.StatusText(http.StatusRequestTimeout))
		}),
	)
}

// endregion
// region - rate limiter

func rateLimiter(lim rate.Limiter) gin.HandlerFunc {
	if lim == nil || lim.Mode() == rate.LimiterModeOff {
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
				switch lim.Mode() {
				case rate.LimiterModeDeny:
					rateLimitDeny(lmtr, c)
				case rate.LimiterModeDelay:
					rateLimitDelay(lmtr, c)
				default:
					c.Next()
				}
			}),
		)
		mw(ctx)
	}
}

func rateLimitDeny(lim *limiter.Limiter, ctx *gin.Context) {
	ctx.Writer.WriteHeader(http.StatusTooManyRequests)
	wait, err := getWait(lim, ctx)
	if err != nil {
		_, _ = ctx.Writer.Write([]byte("Too many requests.\n"))
	} else {
		_, _ = ctx.Writer.Write([]byte(fmt.Sprintf("Too many requests. Try again in %d seconds.\n", wait)))
	}
	ctx.Abort()
}
func rateLimitDelay(lim *limiter.Limiter, ctx *gin.Context) {
	wait, err := getWait(lim, ctx)
	if err != nil {
		_, _ = ctx.Writer.Write([]byte("Too many requests.\n"))
		ctx.Abort()
	}
	timer := time.NewTimer(time.Duration(wait) * time.Second)
	<-timer.C
	ctx.Next()
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
	realIp := ctx.ClientIP() // TODO: use correct way to find "trusted" client ip (see TODO.md)
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

// endregion

//func circuitBreakerMiddleware()

//func realIp() gin.HandlerFunc {
//	return func(ctx *gin.Context) {
//		ctx.Header("X-Real-Ip", ctx.ClientIP())
//	}
//}
