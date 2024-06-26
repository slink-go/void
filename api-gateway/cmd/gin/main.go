package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/api-gateway/cmd/common/variables"
	"github.com/slink-go/api-gateway/discovery"
	"github.com/slink-go/api-gateway/middleware/auth"
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"github.com/slink-go/util/env"
	"github.com/slink-go/util/matcher"
	"github.com/xhit/go-str2duration/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {

	defer func() {
		if err := recover(); err != nil {
			logging.GetLogger("main").Error("%s", err)
		}
	}()

	common.LoadEnv()

	authSkipMatcher = matcher.NewRegexPatternMatcher(env.StringArrayOrEmpty(variables.AuthSkip)...)
	timeoutSkipMatcher = matcher.NewRegexPatternMatcher(env.StringArrayOrEmpty(variables.TimeoutSkip)...)

	gin.SetMode(gin.ReleaseMode)

	var sPort string
	if svcPort := int(env.Int64OrDefault(variables.ServicePort, 0)); svcPort > 0 {
		sPort = fmt.Sprintf(":%d", svcPort)
	}

	var mPort string
	if monPort := int(env.Int64OrDefault(variables.MonitoringPort, 0)); monPort > 0 {
		mPort = fmt.Sprintf(":%d", monPort)
	}

	ec := createEurekaClient()
	dc := createDiscoClient()
	sc := createStaticClient()

	<-startGateway(sPort, mPort, ec, dc, sc)
	time.Sleep(10 * time.Millisecond)
}

func startGateway(proxyAddr, monitoringAddr string, dc ...discovery.Client) chan struct{} {
	reg := registry.NewServiceRegistry(dc...)
	res := resolver.NewServiceResolver(reg)
	proc := resolver.NewPathProcessor()
	ap := createAuthChain()
	udp := createUserDetailsProvider(ap, res, proc)
	pr := createReverseProxy(res, proc)
	limiter := createRateLimiter()
	quitChn := make(chan struct{})
	go NewGinBasedGateway(
		WithAuthProvider(ap),
		WithUserDetailsCache(auth.NewUserDetailsCache(env.DurationOrDefault(variables.AuthCacheTTL, time.Second*30))),
		WithUserDetailsProvider(udp),
		WithRateLimiter(limiter),
		WithReverseProxy(pr),
		WithRegistry(reg),
		WithQuitChn(quitChn),
	).Serve(proxyAddr, monitoringAddr)
	return quitChn
}

func createEurekaClient() discovery.Client {
	if !env.BoolOrDefault(variables.EurekaClientEnabled, false) {
		return nil
	}
	if env.StringOrDefault(variables.EurekaUrl, "") == "" {
		panic("eureka service URL not set")
	}
	dc := discovery.NewEurekaClient(
		discovery.NewEurekaClientConfig().
			WithUrl(env.StringOrDefault(variables.EurekaUrl, "")).
			WithAuth(
				env.StringOrDefault(variables.EurekaLogin, ""),
				env.StringOrDefault(variables.EurekaPassword, ""),
			).
			WithRefresh(env.DurationOrDefault(variables.EurekaRefreshInterval, time.Second*30)).
			WithApplication(env.StringOrDefault(variables.GatewayName, "fiber-gateway")),
	)
	if err := dc.Connect(); err != nil {
		logging.GetLogger("main").Warning("eureka client initialization error: %s", err)
		return nil
	}
	logging.GetLogger("main").Info("started eureka registry")
	return dc
}
func createDiscoClient() discovery.Client {
	if !env.BoolOrDefault(variables.DiscoClientEnabled, false) {
		return nil
	}
	if env.StringOrDefault(variables.DiscoUrl, "") == "" {
		panic("disco service URL not set")
	}
	dc := discovery.NewDiscoClient(
		discovery.NewDiscoClientConfig().
			WithUrl(env.StringOrDefault(variables.DiscoUrl, "")).
			WithBasicAuth(
				env.StringOrDefault(variables.DiscoLogin, ""),
				env.StringOrDefault(variables.DiscoPassword, ""),
			).
			WithApplication(env.StringOrDefault(variables.GatewayName, "fiber-gateway")),
	)

	if err := dc.Connect(make(chan struct{})); err != nil {
		logging.GetLogger("main").Warning("disco client initialization error: %s", err)
		return nil
	}
	logging.GetLogger("main").Info("started disco registry")
	return dc
}
func createStaticClient() discovery.Client {
	filePath := env.StringOrDefault(variables.StaticRegistryFile, "")
	if filePath == "" {
		return nil
	}
	v, err := discovery.LoadFromFile(filePath)
	if err != nil {
		logging.GetLogger("main").Error("static registry initialization error ('%s'): %s", filePath, err)
		return nil
	}
	logging.GetLogger("main").Info("started static registry")
	return v
}

func createAuthChain() security.AuthProvider {
	return security.NewAuthChain(
		security.WithProvider(security.NewHttpHeaderAuthProvider()),
		security.WithProvider(security.NewCookieAuthProvider()),
	)
}
func createUserDetailsProvider(ap security.AuthProvider, res resolver.ServiceResolver, proc resolver.PathProcessor) security.UserDetailsProvider {
	return security.NewTokenBasedUserDetailsProvider(
		security.UdpWithAuthProvider(ap),
		security.UdpWithServiceResolver(res),
		security.UdpWithPathProcessor(proc),
		security.UdpWithMethod(env.StringOrDefault(variables.AuthMethod, "GET")),
		security.UdpWithAuthEndpoint(env.StringOrDefault(variables.AuthEndpoint, "")),
		security.UdpWithResponseParser(security.NewResponseParser(security.WithMappingFile(os.Getenv(variables.AuthResponseMappingFilePath)))),
	)
}
func createReverseProxy(res resolver.ServiceResolver, proc resolver.PathProcessor) *proxy.ReverseProxy {
	return proxy.CreateReverseProxy().WithServiceResolver(res).WithPathProcessor(proc)
}
func createRateLimiter() rate.Limiter {
	var options []rate.Option
	options = append(options, rate.WithLimit(env.Int64OrDefault(variables.LimiterLimit, 10)))
	options = append(options, rate.WithPeriod(env.DurationOrDefault(variables.LimiterPeriod, time.Minute)))
	options = append(options, rate.WithMode(env.StringOrDefault(variables.LimiterMode, "")))
	options = append(options, rate.WithInMemStore())
	options = append(options, parseCustomRateLimits()...)
	limiter := rate.NewLimiter(options...)
	return limiter
}
func parseCustomRateLimits() []rate.Option {
	logger := logging.GetLogger("custom-rate-parser")
	v, err := env.String(variables.LimiterCustomConfig)
	if err != nil {
		return nil
	}
	var result []rate.Option
	for _, part := range strings.Split(v, ",") {
		pattern, period, limit, err := parseCustomRateConfig(part)
		if err != nil {
			logger.Warning("%s", err)
			continue
		}
		logger.Debug("adding custom rate: '%s' '%v' '%v'", pattern, period, limit)
		result = append(result, rate.WithCustom(
			rate.WithCustomPattern(pattern),
			rate.WithCustomLimit(limit),
			rate.WithCustomPeriod(period),
		))
	}
	return result
}

func parseCustomRateConfig(input string) (pattern string, period time.Duration, limit int64, err error) {
	parts := strings.SplitN(input, ":", 3)
	if len(parts) != 3 {
		err = fmt.Errorf("invalid custom config '%s'", input)
		return
	}
	pattern = parts[0]

	v, err := strconv.Atoi(parts[1])
	if err != nil {
		err = fmt.Errorf("could not parse limit '%s': %s", parts[1], err)
		return
	}
	limit = int64(v)

	period, err = str2duration.ParseDuration(strings.ToLower(parts[2]))
	if err != nil {
		err = fmt.Errorf("could not parse duration '%s': %s", parts[2], err)
	}
	return
}
