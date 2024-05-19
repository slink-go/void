package main

import (
	"github.com/gin-gonic/gin"
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/resolver"
)

func main() {

	common.LoadEnv()

	gin.SetMode(gin.ReleaseMode)

	ap := security.NewHttpHeaderAuthProvider()
	udp := security.NewStubUserDetailsProvider()
	//limit := ratelimit.New(10)
	pr := proxy.CreateReverseProxy().
		WithServiceResolver(common.ServiceResolver()).
		WithPathProcessor(resolver.NewPathProcessor())

	go NewGinBasedGateway().
		WithAuthProvider(ap).
		WithUserDetailsProvider(udp).
		//WithRateLimiter().
		WithReverseProxy(pr).
		Serve(":3003")

	go NewGinBasedGateway().
		WithReverseProxy(pr).
		Serve(":3013")

	<-make(chan struct{})

}
