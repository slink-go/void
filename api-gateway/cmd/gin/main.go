package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/discovery"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"time"
)

// TODO:
//		WS Proxy: 	https://stackoverflow.com/questions/73187877/how-do-i-implement-a-wss-reverse-proxy-as-a-gin-route
// 					https://dev.to/hgsgtk/reverse-http-proxy-over-websocket-in-go-part-1-13n4
//					https://gist.github.com/seblegall/2a2697fc56417b24a7ec49eb4a8d7b1b
//  	Monitoring Web UI
//		Needed Middleware
//		CSRF protection (?)
//					https://www.stackhawk.com/blog/golang-csrf-protection-guide-examples-and-how-to-enable-it/
//					https://github.com/utrack/gin-csrf
//		Circuit Breaker:
//					https://gist.github.com/jerryan999/bcfdd746f3f8c2c11c3d27f1b65dfcf3
//					https://pkg.go.dev/github.com/go-kit/kit/circuitbreaker
//					https://medium.com/german-gorelkin/go-patterns-circuit-breaker-921a7489597
//					!!! https://dev.to/he110/circuitbreaker-pattern-in-go-43cn
//					!!! https://github.com/sony/gobreaker
//		Hystrix:
//					https://github.com/afex/hystrix-go
//
//	Middleware:
//		https://github.com/orgs/gin-contrib/repositories
//		https://github.com/gin-gonic/contrib/tree/master/gzip
//

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
		mPort = fmt.Sprintf("127.0.0.1:%d", monPort)
	}

	ec := createEurekaClient()
	dc := createDiscoClient()
	sc := createStaticClient()

	<-startGateway(sPort, mPort, ec, dc, sc)
	time.Sleep(10 * time.Millisecond)
}

func startGateway(proxyAddr, monitoringAddr string, dc ...discovery.Client) chan struct{} {

	ap := security.NewHttpHeaderAuthProvider()
	udp := security.NewStubUserDetailsProvider()
	limiter := rate.NewLimiter(int(env.Int64OrDefault(env.RateLimitRPM, 0)))
	reg := registry.NewServiceRegistry(dc...)

	pr := proxy.CreateReverseProxy().
		WithServiceResolver(resolver.NewServiceResolver(reg)).
		WithPathProcessor(resolver.NewPathProcessor())

	quitChn := make(chan struct{})

	go NewGinBasedGateway().
		WithAuthProvider(ap).
		WithUserDetailsProvider(udp).
		WithRateLimiter(limiter).
		WithReverseProxy(pr).
		WithRegistry(reg).
		WithQuitChn(quitChn).
		Serve(proxyAddr, monitoringAddr)

	return quitChn

}

func createEurekaClient() discovery.Client {
	if env.StringOrDefault(env.EurekaUrl, "") == "" {
		return nil
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
	if env.StringOrDefault(env.DiscoUrl, "") == "" {
		return nil
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
