package main

import (
	"fmt"
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/api-gateway/cmd/common/env"
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

	ap := security.NewHttpHeaderAuthProvider()
	udp := security.NewStubUserDetailsProvider()
	limiter := rate.NewLimiter(10)

	//reg := createStaticRegistry()
	reg := createEurekaRegistry()

	pr := proxy.CreateReverseProxy().
		WithServiceResolver(resolver.NewServiceResolver(reg)).
		WithPathProcessor(resolver.NewPathProcessor())

	base, add := common.GetServicePorts()

	if base > 0 {
		go NewFiberBasedGateway().
			WithAuthProvider(ap).
			WithUserDetailsProvider(udp).
			WithReverseProxy(pr).
			WithRegistry(reg).
			Serve(fmt.Sprintf(":%d", base))
		logging.GetLogger("main").Info(fmt.Sprintf("started api gateway on :%d", base))
	}

	if add > 0 {
		go NewFiberBasedGateway().
			WithAuthProvider(ap).
			WithUserDetailsProvider(udp).
			WithRateLimiter(limiter).
			WithReverseProxy(pr).
			WithRegistry(reg).
			Serve(fmt.Sprintf(":%d", add))
		logging.GetLogger("main").Info(fmt.Sprintf("started api gateway on :%d", add))
	}

	<-make(chan struct{})
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
	panic("implement me")
}
