package main

import (
	"context"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slink-go/logger"
	"github.com/slink-go/logging"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewService(name string) *Service {
	return &Service{
		engine: gin.Default(),
		name:   name,
		logger: logging.GetLogger(name),
	}
}

// TODO: 	embedding https://stackoverflow.com/questions/12536574/can-a-go-struct-inherit-a-set-of-values
//			https://golangbot.com/inheritance/

type Service struct {
	engine                  *gin.Engine
	name                    string
	logger                  logging.Logger
	gracefulShutdownTimeout time.Duration
	quitChn                 chan struct{}
}

func (s *Service) Run(address string) {
	server := &http.Server{
		Addr:    address,
		Handler: s.engine,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Panic("[service][%s] listen: %s\n", address, err)
		}
	}()
	logger.Info("start %s service on %s", s.name, address)
	s.handleBreak(server)
}

func (s *Service) WithMiddleware(middleware ...gin.HandlerFunc) *Service {
	s.engine.Use(middleware...)
	return s
}
func (s *Service) WithOptionalMiddleware(flag bool, middleware ...gin.HandlerFunc) *Service {
	if flag {
		s.engine.Use(middleware...)
	}
	return s
}

func (s *Service) WithGetHandlers(endpoint string, handlerFunc ...gin.HandlerFunc) *Service {
	s.engine.GET(endpoint, handlerFunc...)
	return s
}
func (s *Service) WithPutHandlers(endpoint string, handlerFunc ...gin.HandlerFunc) *Service {
	s.engine.PUT(endpoint, handlerFunc...)
	return s
}
func (s *Service) WithPostHandlers(endpoint string, handlerFunc ...gin.HandlerFunc) *Service {
	s.engine.POST(endpoint, handlerFunc...)
	return s
}
func (s *Service) WithDeleteHandlers(endpoint string, handlerFunc ...gin.HandlerFunc) *Service {
	s.engine.DELETE(endpoint, handlerFunc...)
	return s
}
func (s *Service) WithHeadHandlers(endpoint string, handlerFunc ...gin.HandlerFunc) *Service {
	s.engine.HEAD(endpoint, handlerFunc...)
	return s
}
func (s *Service) WithOptionsHandler(endpoint string, handlerFunc gin.HandlerFunc) *Service {
	s.engine.OPTIONS(endpoint, handlerFunc)
	return s
}
func (s *Service) WithNoRouteHandler(handlerFunc gin.HandlerFunc) *Service {
	s.engine.NoRoute(handlerFunc)
	return s
}

func (s *Service) WithStatic(endpoint, path string) *Service {
	s.engine.Static(endpoint, path)
	return s
}

func (s *Service) WithPrometheus(endpoint ...string) *Service {
	ep := "/prometheus"
	if len(endpoint) > 0 && endpoint[0] != "" {
		ep = endpoint[0]
	}
	return s.WithGetHandlers(ep, gin.WrapH(promhttp.Handler()))
}
func (s *Service) WithProfiler() *Service {
	pprof.Register(s.engine)
	return s
}
func (s *Service) RateLimit(rps int) *Service {
	// TODO: implement me
	return s
}

func (s *Service) WithQuitChn(chn chan struct{}) *Service {
	s.quitChn = chn
	return s
}

func (s *Service) handleBreak(server *http.Server) {
	sigChn := make(chan os.Signal)
	signal.Notify(sigChn, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		switch <-sigChn {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGTERM:
			s.logger.Info("shutdown %s service", s.name)
			if s.quitChn != nil {
				s.quitChn <- struct{}{}
			}
			close(sigChn)
			s.shutdownHttpServer(server)
			if s.quitChn != nil {
				s.quitChn <- struct{}{}
			}
			return
		}
	}
}
func (s *Service) shutdownHttpServer(server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), s.gracefulShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		s.logger.Panic("Server Shutdown:", err)
	}
	select {
	case <-ctx.Done():
		log.Println("graceful shutdown timeout expired")
	}
	log.Println("Server exiting")
}
