package disco

import (
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/cmd/common/util"
	"github.com/slink-go/logging"
	"os"
	"time"
)

type Config struct {
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

func NewConfig() *Config {
	return &Config{
		application: "UNKNOWN",
		hostname:    "",
		port:        int(env.Int64OrDefault(env.ServicePort, 0)),
	}
}

func (c *Config) WithUrl(url string) *Config {
	c.url = url
	return c
}
func (c *Config) WithBasicAuth(username, password string) *Config {
	c.login = username
	c.password = password
	return c
}
func (c *Config) WithApplication(application string) *Config {
	c.application = application
	return c
}
func (c *Config) WithPort(port int) *Config {
	c.port = port
	return c
}
func (c *Config) WithIp(ip string) *Config {
	c.ip = ip
	return c
}
func (c *Config) WithHostname(name string) *Config {
	c.hostname = name
	return c
}
func (c *Config) WithRetry(attempts uint, delay time.Duration) *Config {
	c.retryAttempts = attempts
	c.retryDelay = delay
	return c
}
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.timeout = timeout
	return c
}

func (c *Config) getIP() string {
	if c.ip != "" {
		return c.ip
	}
	return util.GetLocalIP()
}
func (c *Config) getHostname() string {
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
