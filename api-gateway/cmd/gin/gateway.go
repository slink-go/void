package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slink-go/api-gateway/gateway"
	gwctx "github.com/slink-go/api-gateway/middleware/context"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/logging"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type GinBasedGateway struct {
	logger           logging.Logger
	contextProvider  gwctx.Provider
	reverseProxy     *proxy.ReverseProxy
	registry         registry.ServiceRegistry
	proxy            *gin.Engine
	monitoring       *gin.Engine
	monitoringServer *http.Server
	quitChn          chan<- struct{}
	//limiter         fiber.Handler
}

//region - initializers

func NewGinBasedGateway() gateway.Gateway {
	gw := GinBasedGateway{
		logger:          logging.GetLogger("gin-gateway"),
		contextProvider: gwctx.CreateContextProvider(),
	}
	return &gw
}
func (g *GinBasedGateway) WithAuthProvider(ap security.AuthProvider) gateway.Gateway {
	g.contextProvider.WithAuthProvider(ap)
	return g
}
func (g *GinBasedGateway) WithUserDetailsProvider(udp security.UserDetailsProvider) gateway.Gateway {
	g.contextProvider.WithUserDetailsProvider(udp)
	return g
}
func (g *GinBasedGateway) WithReverseProxy(reverseProxy *proxy.ReverseProxy) gateway.Gateway {
	g.reverseProxy = reverseProxy
	return g
}

func (g *GinBasedGateway) WithRateLimiter(limit rate.Limiter) gateway.Gateway {
	//limiterConfig := limiter.Config{
	//	Max:                    limit.GetLimit(),
	//	Expiration:             time.Minute,
	//	SkipFailedRequests:     false,
	//	SkipSuccessfulRequests: false,
	//	LimiterMiddleware:      limiter.FixedWindow{},
	//	LimitReached: func(c *fiber.Ctx) error {
	//		return c.Status(http.StatusTooManyRequests).Send([]byte(fmt.Sprintf("Too Many Requests\n")))
	//	},
	//	KeyGenerator: func(c *fiber.Ctx) string {
	//		proxyIP := c.Get("X-Forwarded-For")
	//		if proxyIP != "" {
	//			return proxyIP
	//		} else {
	//			return c.IP()
	//		}
	//	},
	//}
	//g.limiter = limiter.New(limiterConfig)
	return g
}
func (g *GinBasedGateway) WithRegistry(registry registry.ServiceRegistry) gateway.Gateway {
	g.registry = registry
	return g
}
func (g *GinBasedGateway) WithQuitChn(chn chan struct{}) gateway.Gateway {
	g.quitChn = chn
	return g
}

func (g *GinBasedGateway) Serve(addresses ...string) {
	defer func() {
		if g.quitChn != nil {
			g.quitChn <- struct{}{}
		}
		if err := recover(); err != nil {
			g.logger.Warning("server error: ")
		}
	}()
	if addresses == nil || len(addresses) == 0 {
		panic("service address(es) not set")
	}
	if len(addresses) > 1 && addresses[1] != "" {
		g.startMonitoring(addresses[1])
	}
	//if addresses[0] != "" {
	//	g.startProxyService(addresses[0])
	//} else {
	//	panic("service port not set")
	//}
}

// endregion
//region - monitoring

func (g *GinBasedGateway) startMonitoring(address string) {

	g.monitoring = gin.Default()
	g.setupMonitoringMiddleware()
	//g.setupMonitoringRouteHandlers()
	g.logger.Info("start monitoring service on %s", address)

	g.monitoringServer = &http.Server{
		Addr:    address,
		Handler: g.monitoring,
	}

	go func() {
		if err := g.monitoringServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.logger.Panic("listen: %s\n", err)
		}
	}()

	g.handleBreak("monitoring", g.monitoringServer)

}

func (g *GinBasedGateway) setupMonitoringMiddleware() {
	//p := fiberprometheus.New(fmt.Sprintf("fiber-api-gateway-monitor:%s", address))
	g.monitoring.GET("/prometheus", gin.WrapH(promhttp.Handler()))
	//p.RegisterAt(g.monitoring, "/prometheus")
	//g.monitoring.Use(p.Middleware)
	g.monitoring.Use(gin.Logger())
}

//func (g *GinBasedGateway) setupMonitoringRouteHandlers() {
//	g.monitoring.Get("/", g.monitoringPage)
//	g.monitoring.Static("/s", "./static")
//	g.monitoring.Get("/list", g.listRemotes)
//	g.monitoring.Get("/monitor", monitor.New(monitor.Config{Title: "VOID API Gateway (monitoring)"}))
//}
//func (g *GinBasedGateway) monitoringPage(c *fiber.Ctx) error {
//	c.Set("Content-Type", "text/html")
//	t := templates2.ServicesPage(templates2.Cards(g.registry.List()))
//	err := t.Render(c.Context(), c.Response().BodyWriter())
//	return err
//}
//func (g *GinBasedGateway) listRemotes(c *fiber.Ctx) error {
//	if g.registry != nil {
//		data := g.registry.List()
//		if data == nil || len(data) == 0 {
//			c.Status(fiber.StatusNoContent)
//		} else {
//			buff, err := json.Marshal(data)
//			if err != nil {
//				c.Status(fiber.StatusInternalServerError).Write([]byte(err.Error()))
//			} else {
//				c.Write(buff)
//			}
//		}
//	} else {
//		c.Status(fiber.StatusNoContent)
//	}
//	return nil
//}

// endregion

// region - common

func (g *GinBasedGateway) handleBreak(service string, server *http.Server) {
	sigChn := make(chan os.Signal)
	signal.Notify(sigChn, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		switch <-sigChn {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGTERM:
			g.logger.Info("shutdown %s service", service)
			close(sigChn)
			shutdown(server, g.logger)
			if g.quitChn != nil {
				g.quitChn <- struct{}{}
			}
			return
		}
	}
}

func shutdown(server *http.Server, logger logging.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Panic("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}

// endregion
