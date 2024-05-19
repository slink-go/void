package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/palantir/stacktrace"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"go.uber.org/ratelimit"
	"net/http"
	"os"
	"strings"
)

var limit ratelimit.Limiter
var logger logging.Logger
var reverseProxy *proxy.ReverseProxy

func leakBucket() gin.HandlerFunc {
	//prev := time.Now()
	return func(ctx *gin.Context) {
		limit.Take()
		//now := limit.Take()
		//logger.Info("%v", now.Sub(prev))
		//prev = now
	}
}

func main() {
	loadEnv()

	limit = ratelimit.New(10)

	reverseProxy = proxy.CreateReverseProxy().
		WithServiceResolver(serviceResolver()).
		WithPathProcessor(resolver.NewPathProcessor())

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(leakBucket())
	router.GET("*path", proxyHandler)
	router.POST("*path", proxyHandler)
	router.Run(":3003")
}

func proxyHandler(ctx *gin.Context) {

	defer func() {
		if err := recover(); err != nil {
			logger.Warning("panic: %v", err)
		}
	}()

	target, err := reverseProxy.ResolveTarget(ctx.Request.URL.Path)
	if err != nil {
		logger.Warning("%s", stacktrace.RootCause(err))
		switch err.(type) {
		case *resolver.ErrEmptyBaseUrl:
			ctx.AbortWithStatus(http.StatusBadGateway)
		case *resolver.ErrInvalidPath:
			ctx.AbortWithStatus(http.StatusBadRequest)
		case *resolver.ErrServiceUnavailable:
			ctx.AbortWithStatus(http.StatusServiceUnavailable)
		}
		return
	}
	logger.Info("resolved url: %s://%s%s%s -> %s", "http", ctx.Request.Host, ctx.Request.URL.Path, queryParams(ctx), target)

	// TODO: implement it
	//headers, err := preprocessRequest(ctx)
	//if err != nil {
	//	ctx.AbortWithStatus(http.StatusUnauthorized)
	//}
	headers := make(map[string][]string)
	headers["Header"] = []string{"STUB"}
	reverseProxy.Proxy(target, headers).ServeHTTP(ctx.Writer, ctx.Request)
}

func serviceRegistry() resolver.ServiceRegistry {
	var registry = make(map[string][]string, 2)
	registry["service-a"] = []string{"localhost:3101", "localhost:3102", "localhost:3103"}
	registry["service-b"] = []string{"localhost:3201", "localhost:3202", "localhost:3203"}
	return resolver.NewStaticServiceRegistry(registry)
}
func serviceResolver() resolver.ServiceResolver {
	return resolver.NewServiceResolver(serviceRegistry())
}
func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		os.Setenv("GO_ENV", "dev")
		logging.GetLogger("main").Warning("could not read config from .env file")
	}
	logger = logging.GetLogger("main")
}

func queryParams(ctx *gin.Context) string {
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

//func preprocessRequest(ctx *gin.Context) (map[string][]string, error) {
//	headers := map[string][]string(ctx.Request.Header)
//
//	userId, userName, userRole, err := auth(h.Get(headers[h.CtxAuthToken]))
//
//	if err != nil {
//		return nil, err
//	}
//
//	locale := getRequestLocale(ctx.Request)
//
//	headers[h.CtxLocale] = []string{locale}
//	headers[h.CtxUserId] = []string{userId}
//	headers[h.CtxUserName] = []string{userName}
//	headers[h.CtxUserRole] = []string{userRole}
//
//	return headers, nil
//}

//func auth(token string) (id, name, role string, err error) {
//	return "1", "Vasya", "user", nil
//}

//func getRequestLocale(request *http.Request) string {
//	return "ru"
//}
