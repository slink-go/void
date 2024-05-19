package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	h "github.com/slink-go/api-gateway/middleware/context"
	"github.com/slink-go/logging"
	"strings"
	"time"
)

type Service struct {
	logger    logging.Logger
	serviceId string
}

func Create(serviceId, boundAddress string) *Service {
	service := Service{
		logger:    logging.GetLogger(fmt.Sprintf("%s-%s", "service", strings.ToLower(serviceId))),
		serviceId: serviceId,
	}

	app := fiber.New()
	app.Get("/", service.rootHandler)
	app.Get("/api/test", service.testHandler)
	app.Get("/api/slow", service.slowHandler)
	app.Listen(boundAddress)

	return &service
}

func (s *Service) slowHandler(c *fiber.Ctx) error {
	s.logger.Info("[slow] start %s", c.Context().RemoteAddr())
	time.Sleep(3 * time.Second)
	err := c.SendString(
		fmt.Sprintf("SLOW %s\n", s.serviceId),
	)
	s.logger.Info("[slow] complete %s", c.Context().RemoteAddr())
	return err
}
func (s *Service) testHandler(c *fiber.Ctx) error {
	s.logger.Info("%s %v '%v'", c.Context().RemoteAddr(), c.GetReqHeaders(), s.queryParams(c))
	err := c.SendString(
		fmt.Sprintf(
			"TEST %s\n%s\n%s\n",
			s.serviceId,
			h.GetHeader(c.GetReqHeaders()["Header"]),
			c.Query("param"),
		),
	)
	return err
}
func (s *Service) rootHandler(c *fiber.Ctx) error {
	s.logger.Trace("%s %v\n", c.Context().RemoteAddr(), c.GetReqHeaders())
	err := c.SendString(
		fmt.Sprintf(
			"Hello from service %s!\n(%s, %s)\n",
			s.serviceId,
			h.GetHeader(c.GetReqHeaders()[h.CtxAuthToken]),
			c.Query("key"),
		),
	)
	return err
}
func (s *Service) queryParams(c *fiber.Ctx) string {
	result := ""
	for p, v := range c.Queries() {
		result = result + p
		result = result + "="
		result = result + v
		result = result + ", "
	}
	return strings.TrimSuffix(result, ", ")
}
