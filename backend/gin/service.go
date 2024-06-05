package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/discovery"
	h "github.com/slink-go/api-gateway/middleware/constants"
	"github.com/slink-go/logging"
	"math/rand/v2"
	"net/http"
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
	stream        *Stream
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
		stream:        NewStreamingServer(),
	}
	return &service
}
func (s *Service) Start() {

	if s.discovery != nil {
		if err := s.discovery.Connect(); err != nil {
			logging.GetLogger("--").Warning("%s", err)
		}
	}

	go s.StartDataGenerator(s.stream)

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.GET("/", s.rootHandler)
	router.GET("/api/test", s.testHandler)
	router.GET("/api/slow", s.slowHandler)
	router.GET("/api/apps", s.appsHandler)

	router.GET("/api/sse", SseHeadersMiddleware(), s.sseClientConnectionMiddleware(), s.sseStreamHandler)
	router.GET("/api/ws", WsUpgraderMiddleware(), s.wsClientConnectionMiddleware(), s.wsStreamHandler)

	router.Run(s.address)

	//engine := html.New("./views", ".html")
	//
	//app := fiber.New(fiber.Config{
	//	Views: engine,
	//})
	//
	//app.Static("/s", "./static/")
	//app.Use(pprof.New())
	//app.Get("/", s.rootHandler)
	//
	//app.Get("/api/slow", s.slowHandler)
	//app.Get("/api/apps", s.appsListHandler)
	//
	//app.Get("/sse-test", s.sseTestHandler)
	//app.Get("/api/sse", s.sseHandler)
	//
	//app.Use("/ws", func(c *fiber.Ctx) error {
	//	// IsWebSocketUpgrade returns true if the client
	//	// requested upgrade to the WebSocket protocol.
	//	if websocket.IsWebSocketUpgrade(c) {
	//		c.Locals("allowed", true)
	//		return c.Next()
	//	}
	//	return fiber.ErrUpgradeRequired
	//})
	//app.Get("/ws/:id", websocket.New(s.wsHandler))
	//
	//s.StartDataGenerator()
	//
	//app.Listen(s.address)
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
	url := env.StringOrDefault(env.DiscoUrl, "")
	lg := env.StringOrDefault(env.DiscoLogin, "")
	pw := env.StringOrDefault(env.DiscoPassword, "")
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
	url := env.StringOrDefault(env.EurekaUrl, "")
	lg := env.StringOrDefault(env.EurekaLogin, "")
	pw := env.StringOrDefault(env.EurekaPassword, "")
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
			WithHeartBeat(env.DurationOrDefault(env.EurekaHeartbeatInterval, time.Second*10)).
			WithRefresh(env.DurationOrDefault(env.EurekaRefreshInterval, time.Second*30)).
			WithApplication(applicationId).
			//WithInstanceId(fmt.Sprintf("%s-%s", applicationId, instanceId)).
			WithPort(port),
	)
}

// endregion
// region - basic

func (s *Service) rootHandler(c *gin.Context) {
	s.logger.Trace("%s %v\n", c.RemoteIP(), c.Request.Header)
	c.String(http.StatusOK,
		"Hello from service %s-%s!\n(%s, %s)\n",
		s.applicationId,
		s.instanceId,
		c.Request.Header.Get(h.CtxAuthToken),
		c.Query("key"),
	)
}
func (s *Service) testHandler(c *gin.Context) {
	s.logger.Info("%s %v '%v'", c.RemoteIP, c.Request.Header, s.getQueryParams(c))
	c.String(
		http.StatusOK,
		"TEST %s-%s\nBOUND:%s\nHEADERS: %s\nQUERY PARAMS: %s\n",
		s.applicationId,
		s.instanceId,
		s.address,
		s.getHeaders(c),
		s.getQueryParams(c),
	)
}
func (s *Service) slowHandler(c *gin.Context) {
	s.logger.Info("[slow] start %s", c.RemoteIP())
	seconds := time.Duration(rand.IntN(5)) * time.Second
	time.Sleep(seconds)
	c.String(
		http.StatusOK,
		"SLOW %s-%s (%s)\n",
		s.applicationId,
		s.instanceId,
		seconds,
	)
	s.logger.Info("[slow] complete %s", c.RemoteIP())
}
func (s *Service) appsHandler(c *gin.Context) {
	remotes := s.discovery.Services()
	if remotes == nil {
		c.String(http.StatusNoContent, "no remotes discovered")
		return
	}
	for _, v := range remotes.List() {
		s.logger.Info("remote: %s", v)
	}
	c.JSON(http.StatusOK, remotes.List())
}

// endregion
// region - helpers

func (s *Service) getHeaders(c *gin.Context) string {
	var headers []string
	for k, _ := range c.Request.Header {
		headers = append(headers, k)
	}
	slices.Sort(headers)
	var result []string
	for _, h := range headers {
		list, ok := c.Request.Header[h]
		if ok {
			for _, v := range list {
				result = append(result, fmt.Sprintf("%s=%s", h, v))
			}
		}
	}
	return strings.Join(result, ", ")
}
func (s *Service) getQueryParams(c *gin.Context) string {
	var result []string
	for p, list := range c.Request.URL.Query() {
		for _, v := range list {
			result = append(result, fmt.Sprintf("%s=%s", p, v))
		}
	}
	return strings.Join(result, ", ")
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
