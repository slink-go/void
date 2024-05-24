package main

import (
	"fmt"
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

// https://docs.gofiber.io/category/-middleware/

func main() {

	defer func() {
		if err := recover(); err != nil {
			logging.GetLogger("main").Error("%s", err)
		}
	}()

	common.LoadEnv()

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

	startGateway("common", sPort, mPort, ec, dc, sc)

	<-make(chan struct{})
}

func startGateway(title, saddr, maddr string, dc ...discovery.Client) {

	ap := security.NewHttpHeaderAuthProvider()
	udp := security.NewStubUserDetailsProvider()
	limiter := rate.NewLimiter(int(env.Int64OrDefault(env.RateLimitRPM, 0)))
	reg := registry.NewServiceRegistry(dc...)

	pr := proxy.CreateReverseProxy().
		WithServiceResolver(resolver.NewServiceResolver(reg)).
		WithPathProcessor(resolver.NewPathProcessor())

	go NewFiberBasedGateway().
		WithAuthProvider(ap).
		WithUserDetailsProvider(udp).
		WithRateLimiter(limiter).
		WithReverseProxy(pr).
		WithRegistry(reg).
		Serve(saddr, maddr)

	//logging.GetLogger("main").Info(fmt.Sprintf("started %s api gateway on %d", title, port))

}

func createEurekaClient() discovery.Client {
	dc := discovery.NewEurekaClient(
		discovery.NewEurekaClientConfig().
			WithUrl(env.StringOrDefault(env.EurekaUrl, "")).
			WithAuth(
				env.StringOrDefault(env.EurekaLogin, ""),
				env.StringOrDefault(env.EurekaPassword, ""),
			).
			WithRefresh(env.DurationOrDefault(env.EurekaRefreshInterval, time.Second*30)).
			WithApplication("fiber-gateway"),
	)
	dc.Connect()
	return dc
}
func createDiscoClient() discovery.Client {
	dc := discovery.NewDiscoClient(
		discovery.NewDiscoClientConfig().
			WithUrl(env.StringOrDefault(env.DiscoUrl, "")).
			WithBasicAuth(
				env.StringOrDefault(env.DiscoLogin, ""),
				env.StringOrDefault(env.DiscoPassword, ""),
			).
			WithApplication("fiber-gateway"),
	)
	dc.Connect()
	return dc
}
func createStaticClient() discovery.Client {
	filePath := env.StringOrDefault(env.StaticRegistryFile, "")
	v, err := discovery.LoadFromFile(filePath)
	if err != nil {
		logging.GetLogger("main").Error("static registry initialization error ('%s'): %s", filePath, err)
		return nil
	}
	return v
}
