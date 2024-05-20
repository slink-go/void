package eureka

import (
	"fmt"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"time"
)

func NewConfig() *Config {
	return &Config{
		fetch:       false,
		register:    false,
		application: "UNKNOWN",
		hostname:    "localhost",
		port:        int(env.Int64OrDefault(env.ServicePort, 0)),
	}
}

type Config struct {
	url               string // eureka server url
	login             string // [optional] eureka server access login
	password          string // [optional] eureka server access password
	fetch             bool   // [false] should client fetch service registry from eureka
	register          bool   // [false] should client register itself on eureka server
	heartBeatInterval time.Duration
	refreshInterval   time.Duration
	application       string // application name
	instanceId        string
	port              int
	ip                string
	hostname          string
}

func (c *Config) WithUrl(url string) *Config {
	c.url = url
	return c
}
func (c *Config) WithAuth(login, password string) *Config {
	c.login = login
	c.password = password
	return c
}
func (c *Config) WithHeartBeat(interval time.Duration) *Config {
	c.register = true
	c.heartBeatInterval = interval
	return c
}
func (c *Config) WithRefresh(interval time.Duration) *Config {
	c.fetch = true
	c.refreshInterval = interval
	return c
}
func (c *Config) WithApplication(name string) *Config {
	c.application = name
	return c
}
func (c *Config) WithInstanceId(id string) *Config {
	c.instanceId = id
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

func (c *Config) getInstanceId() string {
	if c.instanceId != "" {
		return c.instanceId
	}
	return fmt.Sprintf("%s:%s:%d", c.getIP(), c.application, int(env.Int64OrDefault(env.ServicePort, 0)))
}
func (c *Config) getIP() string {
	if c.ip != "" {
		return c.ip
	}
	return GetLocalIP()
}
