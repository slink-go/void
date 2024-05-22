package main

import (
	"fmt"
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/discovery/disco"
	"github.com/slink-go/api-gateway/discovery/eureka"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"time"
)

// https://docs.gofiber.io/category/-middleware/

func main() {

	defer func() {
		if err := recover(); err != nil {
			logging.GetLogger("main").Error("%s", err)
		}
	}()

	common.LoadEnv()

	port := int(env.Int64OrDefault(env.ServicePort, 0))

	if port > 0 {
		startEurekaGateway(port)
		startDiscoGateway(port + 1)
	} else {
		panic("service port not set")
	}

	<-make(chan struct{})
}

func startEurekaGateway(port int) {

	ap := security.NewHttpHeaderAuthProvider()
	udp := security.NewStubUserDetailsProvider()
	limiter := rate.NewLimiter(int(env.Int64OrDefault(env.RateLimitRPM, 0)))
	reg := createEurekaRegistry()

	pr := proxy.CreateReverseProxy().
		WithServiceResolver(resolver.NewServiceResolver(reg)).
		WithPathProcessor(resolver.NewPathProcessor())

	go NewFiberBasedGateway().
		WithAuthProvider(ap).
		WithUserDetailsProvider(udp).
		WithRateLimiter(limiter).
		WithReverseProxy(pr).
		WithRegistry(reg).
		Serve(fmt.Sprintf(":%d", port))

	logging.GetLogger("main").Info(fmt.Sprintf("started eureka-based api gateway on :%d", port))

}
func startDiscoGateway(port int) {

	ap := security.NewHttpHeaderAuthProvider()
	udp := security.NewStubUserDetailsProvider()
	limiter := rate.NewLimiter(int(env.Int64OrDefault(env.RateLimitRPM, 0)))
	reg := createDiscoRegistry()

	pr := proxy.CreateReverseProxy().
		WithServiceResolver(resolver.NewServiceResolver(reg)).
		WithPathProcessor(resolver.NewPathProcessor())

	go NewFiberBasedGateway().
		WithAuthProvider(ap).
		WithUserDetailsProvider(udp).
		WithRateLimiter(limiter).
		WithReverseProxy(pr).
		WithRegistry(reg).
		Serve(fmt.Sprintf(":%d", port))

	logging.GetLogger("main").Info(fmt.Sprintf("started disco-based api gateway on :%d", port))

}

func createStaticRegistry() registry.ServiceRegistry {
	return registry.NewStaticRegistry(common.Services())
}
func createEurekaRegistry() registry.ServiceRegistry {
	eurekaClientConfig := eureka.Config{}
	eurekaClientConfig.
		WithUrl(env.StringOrDefault(env.EurekaUrl, "")).
		WithAuth(
			env.StringOrDefault(env.EurekaLogin, ""),
			env.StringOrDefault(env.EurekaPassword, ""),
		).
		WithRefresh(env.DurationOrDefault(env.EurekaRefreshInterval, time.Second*30)).
		WithApplication("fiber-gateway")
	dc := eureka.NewEurekaClient(&eurekaClientConfig) // eureka discovery client
	dc.Connect()
	return registry.NewDiscoveryRegistry(dc)
}
func createDiscoRegistry() registry.ServiceRegistry {
	discoClientConfig := disco.Config{}
	discoClientConfig.
		WithUrl(env.StringOrDefault(env.DiscoUrl, "")).
		WithBasicAuth(
			env.StringOrDefault(env.DiscoLogin, ""),
			env.StringOrDefault(env.DiscoPassword, ""),
		).
		WithApplication("fiber-gateway")
	dc := disco.NewDiscoClient(&discoClientConfig) // disco discovery client
	dc.Connect()
	return registry.NewDiscoveryRegistry(dc)
}
