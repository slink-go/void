package main

import (
	"fmt"
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
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

	pr := proxy.CreateReverseProxy().
		WithServiceResolver(common.ServiceResolver()).
		WithPathProcessor(resolver.NewPathProcessor())

	base, add := common.GetServicePorts()

	if base > 0 {
		go NewFiberBasedGateway().
			WithAuthProvider(ap).
			WithUserDetailsProvider(udp).
			WithReverseProxy(pr).
			Serve(fmt.Sprintf(":%d", base))
		logging.GetLogger("main").Info(fmt.Sprintf("started api gateway on :%d", base))
	}

	if add > 0 {
		go NewFiberBasedGateway().
			WithAuthProvider(ap).
			WithUserDetailsProvider(udp).
			WithRateLimiter(limiter).
			WithReverseProxy(pr).
			Serve(fmt.Sprintf(":%d", add))
		logging.GetLogger("main").Info(fmt.Sprintf("started api gateway on :%d", add))
	}

	<-make(chan struct{})
}
