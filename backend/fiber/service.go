package main

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/template/html/v2"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/cmd/common/variables"
	"github.com/slink-go/api-gateway/discovery"
	h "github.com/slink-go/api-gateway/middleware/constants"
	"github.com/slink-go/logging"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	logger        logging.Logger
	applicationId string
	instanceId    string
	discovery     discovery.Client
	address       string
	sseChannels   map[chan string]struct{}
	sseData       chan string
}

func Create(applicationId, instanceId, boundAddress, discoveryType string) *Service {
	dc := initDiscoveryClient(applicationId, instanceId, boundAddress, discoveryType)
	if dc == nil {
		return nil
	}

	service := Service{
		logger: logging.GetLogger(
			fmt.Sprintf(
				"%s-%s-%s",
				"service",
				strings.ToLower(applicationId),
				strings.ToLower(instanceId),
			),
		),
		address:       boundAddress,
		applicationId: applicationId,
		instanceId:    instanceId,
		discovery:     dc,
		sseChannels:   make(map[chan string]struct{}),
	}
	return &service
}
func (s *Service) Start() {

	if s.discovery != nil {
		if err := s.discovery.Connect(); err != nil {
			logging.GetLogger("--").Warning("%s", err)
		}
	}

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Static("/s", "./static/")
	app.Use(pprof.New())
	app.Get("/", s.rootHandler)

	app.Get("/api/test", s.testHandler)
	app.Get("/api/slow", s.slowHandler)
	app.Get("/api/apps", s.appsListHandler)

	app.Get("/sse-test", s.sseTestHandler)
	app.Get("/api/sse", s.sseHandler)

	s.StartDataGenerator()

	app.Listen(s.address)
}

//region - discovery

func initDiscoveryClient(applicationId, instanceId, boundAddress, discoveryService string) discovery.Client {
	switch discoveryService {
	case "eureka":
		return initEurekaClient(applicationId, instanceId, boundAddress)
	case "disco":
		return initDiscoClient(applicationId, boundAddress)
	default:
		return nil
		//panic(fmt.Errorf("unsupported discovery service: %s", discoveryService))
	}
}
func initDiscoClient(applicationId, boundAddress string) discovery.Client {
	url := env.StringOrDefault(variables.DiscoUrl, "")
	lg := env.StringOrDefault(variables.DiscoLogin, "")
	pw := env.StringOrDefault(variables.DiscoPassword, "")
	if url == "" {
		return nil
	}
	logging.GetLogger("backend").Warning("register on disco (%s)", url)

	port, err := strconv.Atoi(strings.Split(boundAddress, ":")[1])
	if err != nil {
		port = 0
	}

	return discovery.NewDiscoClient(
		discovery.NewDiscoClientConfig().
			WithUrl(url).
			WithBasicAuth(lg, pw).
			WithApplication(applicationId).
			WithPort(port),
	)
}
func initEurekaClient(applicationId, instanceId, boundAddress string) discovery.Client {
	url := env.StringOrDefault(variables.EurekaUrl, "")
	lg := env.StringOrDefault(variables.EurekaLogin, "")
	pw := env.StringOrDefault(variables.EurekaPassword, "")
	if url == "" {
		return nil
	}
	logging.GetLogger("backend").Warning("register on eureka (%s)", url)

	port, err := strconv.Atoi(strings.Split(boundAddress, ":")[1])
	if err != nil {
		port = 0
	}

	return discovery.NewEurekaClient(
		discovery.NewEurekaClientConfig().
			WithUrl(url).
			WithAuth(lg, pw).
			WithHeartBeat(env.DurationOrDefault(variables.EurekaHeartbeatInterval, time.Second*10)).
			WithRefresh(env.DurationOrDefault(variables.EurekaRefreshInterval, time.Second*30)).
			WithApplication(applicationId).
			//WithInstanceId(fmt.Sprintf("%s-%s", applicationId, instanceId)).
			WithPort(port),
	)
}

// endregion
// region - basic

func (s *Service) rootHandler(c *fiber.Ctx) error {
	s.logger.Trace("%s %v\n", c.Context().RemoteAddr(), c.GetReqHeaders())
	err := c.SendString(
		fmt.Sprintf(
			"Hello from service %s-%s!\n(%s, %s)\n",
			s.applicationId,
			s.instanceId,
			h.GetHeader(c.GetReqHeaders()[h.CtxAuthToken]),
			c.Query("key"),
		),
	)
	return err
}
func (s *Service) appsListHandler(c *fiber.Ctx) error {
	svcs := s.discovery.Services()
	buff, err := json.Marshal(svcs.List())
	if err != nil {
		return err
	}
	_, err = c.Write(buff)
	return err
}
func (s *Service) slowHandler(c *fiber.Ctx) error {
	s.logger.Info("[slow] start %s", c.Context().RemoteAddr())
	time.Sleep(3 * time.Second)
	err := c.SendString(
		fmt.Sprintf("SLOW %s-%s\n", s.applicationId, s.instanceId),
	)
	s.logger.Info("[slow] complete %s", c.Context().RemoteAddr())
	return err
}
func (s *Service) testHandler(c *fiber.Ctx) error {
	s.logger.Info("%s %v '%v'", c.Context().RemoteAddr(), c.GetReqHeaders(), s.queryParams(c))
	err := c.SendString(
		fmt.Sprintf(
			"TEST %s-%s\nBOUND:%s\nHEADERS: %s\nQUERY PARAMS: %s\n",
			s.applicationId,
			s.instanceId,
			s.address,
			s.getHeaders(c),
			s.getQueryParams(c),
		),
	)
	return err
}

// endregion
// region - helpers

func (s *Service) getHeaders(c *fiber.Ctx) string {
	var headers []string
	for k, _ := range c.GetReqHeaders() {
		headers = append(headers, k)
	}
	slices.Sort(headers)
	var result []string
	for _, h := range headers {
		list, ok := c.GetReqHeaders()[h]
		if ok {
			for _, v := range list {
				result = append(result, fmt.Sprintf("%s=%s", h, v))
			}
		}
	}
	return strings.Join(result, ", ")
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
func (s *Service) getQueryParams(c *fiber.Ctx) string {
	return strings.Join(
		strings.Split(
			string(c.Request().URI().QueryString()),
			"&",
		),
		", ",
	)
}

// endregion

// WithUrl(url string)
// WithAuth(login, password string)
// WithHeartBeat(interval time.Duration)
// WithRefresh(interval time.Duration)
// WithApplication(name string)
// WithInstanceId(id string)
// WithPort(port int)
// WithIp(ip string)
