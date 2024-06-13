package discovery

import (
	"github.com/slink-go/api-gateway/cmd/common/variables"
	"github.com/slink-go/api-gateway/discovery/util"
	"github.com/slink-go/logging"
	"github.com/slink-go/util/env"
	"os"
	"time"
)

type discoConfig struct {
	url           string // disco server url
	login         string // [optional] eureka server access login
	password      string // [optional] eureka server access password
	application   string // application name
	port          int    // client application port
	ip            string // client application ip
	hostname      string // client application hostname
	retryAttempts uint
	retryDelay    time.Duration
	timeout       time.Duration
}

func NewDiscoClientConfig() *discoConfig {
	return &discoConfig{
		application: "UNKNOWN",
		hostname:    "",
		port:        int(env.Int64OrDefault(variables.ServicePort, 0)),
	}
}

func (c *discoConfig) WithUrl(url string) *discoConfig {
	c.url = url
	return c
}
func (c *discoConfig) WithBasicAuth(username, password string) *discoConfig {
	c.login = username
	c.password = password
	return c
}
func (c *discoConfig) WithApplication(application string) *discoConfig {
	c.application = application
	return c
}
func (c *discoConfig) WithPort(port int) *discoConfig {
	c.port = port
	return c
}
func (c *discoConfig) WithIp(ip string) *discoConfig {
	c.ip = ip
	return c
}
func (c *discoConfig) WithHostname(name string) *discoConfig {
	c.hostname = name
	return c
}
func (c *discoConfig) WithRetry(attempts uint, delay time.Duration) *discoConfig {
	c.retryAttempts = attempts
	c.retryDelay = delay
	return c
}
func (c *discoConfig) WithTimeout(timeout time.Duration) *discoConfig {
	c.timeout = timeout
	return c
}

func (c *discoConfig) getIP() string {
	if c.ip != "" {
		return c.ip
	}
	return util.GetLocalIP()
}
func (c *discoConfig) getHostname() string {
	if c.hostname == "" {
		v, err := os.Hostname()
		if err != nil {
			logging.GetLogger("disco-config").Warning("could not get hostname: %s", err)
			return "localhost"
		}
		return v
	}
	return c.hostname
}
