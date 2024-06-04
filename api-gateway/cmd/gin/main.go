package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/discovery"
	"github.com/slink-go/api-gateway/middleware/auth"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"os"
	"time"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			logging.GetLogger("main").Error("%s", err)
		}
	}()

	common.LoadEnv()

	gin.SetMode(gin.ReleaseMode)

	var sPort string
	if svcPort := int(env.Int64OrDefault(env.ServicePort, 0)); svcPort > 0 {
		sPort = fmt.Sprintf(":%d", svcPort)
	}

	var mPort string
	if monPort := int(env.Int64OrDefault(env.MonitoringPort, 0)); monPort > 0 {
		mPort = fmt.Sprintf(":%d", monPort)
	}

	ec := createEurekaClient()
	dc := createDiscoClient()
	sc := createStaticClient()

	<-startGateway(sPort, mPort, ec, dc, sc)
	time.Sleep(10 * time.Millisecond)
}

func startGateway(proxyAddr, monitoringAddr string, dc ...discovery.Client) chan struct{} {
	reg := registry.NewServiceRegistry(dc...)
	res := resolver.NewServiceResolver(reg)
	proc := resolver.NewPathProcessor()
	ap := createAuthChain()
	udp := createUserDetailsProvider(ap, res, proc)
	pr := createReverseProxy(res, proc)
	limiter := createRateLimiter()
	quitChn := make(chan struct{})
	go NewGinBasedGateway().
		WithAuthProvider(ap).
		WithUserDetailsCache(auth.NewUserDetailsCache(env.DurationOrDefault(env.AuthCacheTTL, time.Second*30))).
		WithUserDetailsProvider(udp).
		WithRateLimiter(limiter).
		WithReverseProxy(pr).
		WithRegistry(reg).
		WithQuitChn(quitChn).
		Serve(proxyAddr, monitoringAddr)
	return quitChn
}

func createEurekaClient() discovery.Client {
	if !env.BoolOrDefault(env.EurekaClientEnabled, false) {
		return nil
	}
	if env.StringOrDefault(env.EurekaUrl, "") == "" {
		panic("eureka service URL not set")
	}
	dc := discovery.NewEurekaClient(
		discovery.NewEurekaClientConfig().
			WithUrl(env.StringOrDefault(env.EurekaUrl, "")).
			WithAuth(
				env.StringOrDefault(env.EurekaLogin, ""),
				env.StringOrDefault(env.EurekaPassword, ""),
			).
			WithRefresh(env.DurationOrDefault(env.EurekaRefreshInterval, time.Second*30)).
			WithApplication(env.StringOrDefault(env.GatewayName, "fiber-gateway")),
	)
	if err := dc.Connect(); err != nil {
		logging.GetLogger("main").Warning("eureka client initialization error: %s", err)
		return nil
	}
	logging.GetLogger("main").Info("started eureka registry")
	return dc
}
func createDiscoClient() discovery.Client {
	if !env.BoolOrDefault(env.DiscoClientEnabled, false) {
		return nil
	}
	if env.StringOrDefault(env.DiscoUrl, "") == "" {
		panic("disco service URL not set")
	}
	dc := discovery.NewDiscoClient(
		discovery.NewDiscoClientConfig().
			WithUrl(env.StringOrDefault(env.DiscoUrl, "")).
			WithBasicAuth(
				env.StringOrDefault(env.DiscoLogin, ""),
				env.StringOrDefault(env.DiscoPassword, ""),
			).
			WithApplication(env.StringOrDefault(env.GatewayName, "fiber-gateway")),
	)
	if err := dc.Connect(); err != nil {
		logging.GetLogger("main").Warning("disco client initialization error: %s", err)
		return nil
	}
	logging.GetLogger("main").Info("started disco registry")
	return dc
}
func createStaticClient() discovery.Client {
	filePath := env.StringOrDefault(env.StaticRegistryFile, "")
	if filePath == "" {
		return nil
	}
	v, err := discovery.LoadFromFile(filePath)
	if err != nil {
		logging.GetLogger("main").Error("static registry initialization error ('%s'): %s", filePath, err)
		return nil
	}
	logging.GetLogger("main").Info("started static registry")
	return v
}

func createAuthChain() security.AuthProvider {
	return security.NewAuthChain().
		WithProvider(security.NewHttpHeaderAuthProvider()).
		WithProvider(security.NewCookieAuthProvider())
}
func createUserDetailsProvider(ap security.AuthProvider, res resolver.ServiceResolver, proc resolver.PathProcessor) security.UserDetailsProvider {
	return security.NewTokenBasedUserDetailsProvider().
		WithAuthProvider(ap).
		WithServiceResolver(res).
		WithPathProcessor(proc).
		WithMethod(env.StringOrDefault(env.AuthMethod, "GET")).
		WithAuthEndpoint(env.StringOrDefault(env.AuthEndpoint, "")).
		WithResponseParser(security.NewResponseParser().LoadMapping(os.Getenv(env.AuthResponseMappingFilePath)))
}
func createReverseProxy(res resolver.ServiceResolver, proc resolver.PathProcessor) *proxy.ReverseProxy {
	return proxy.CreateReverseProxy().WithServiceResolver(res).WithPathProcessor(proc)
}
func createRateLimiter() rate.Limiter {
	limiter := rate.NewLimiter(
		rate.WithLimit(env.Int64OrDefault(env.LimiterRateLimit, 10)),
		rate.WithPeriod(env.DurationOrDefault(env.LimiterPeriod, time.Minute)),
		// TODO: implement configurable custom limits
		rate.WithCustom(
			rate.WithCustomPattern("*/api/*"),
			rate.WithCustomLimit(5),
			rate.WithCustomPeriod(time.Second*5),
		),
		rate.WithCustom(
			rate.WithCustomPattern("*/api/test/*"),
			rate.WithCustomLimit(15),
			rate.WithCustomPeriod(time.Second),
		),
		rate.WithInMemStore(),
	)
	return limiter
}
