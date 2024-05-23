package discovery

import (
	"fmt"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/discovery/util"
	"github.com/slink-go/logging"
	"os"
	"time"
)

func NewEurekaClientConfig() *eurekaConfig {
	return &eurekaConfig{
		fetch:       false, // disable registry fetching by default
		register:    false, // disable registration on eureka by default
		application: "UNKNOWN",
		hostname:    "",
		port:        int(env.Int64OrDefault(env.ServicePort, 0)),
	}
}

type eurekaConfig struct {
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

func (c *eurekaConfig) WithUrl(url string) *eurekaConfig {
	c.url = url
	return c
}
func (c *eurekaConfig) WithAuth(login, password string) *eurekaConfig {
	c.login = login
	c.password = password
	return c
}
func (c *eurekaConfig) WithHeartBeat(interval time.Duration) *eurekaConfig {
	c.register = true
	c.heartBeatInterval = interval
	return c
}
func (c *eurekaConfig) WithRefresh(interval time.Duration) *eurekaConfig {
	c.fetch = true
	c.refreshInterval = interval
	return c
}
func (c *eurekaConfig) WithApplication(name string) *eurekaConfig {
	c.application = name
	return c
}
func (c *eurekaConfig) WithInstanceId(id string) *eurekaConfig {
	c.instanceId = id
	return c
}
func (c *eurekaConfig) WithPort(port int) *eurekaConfig {
	c.port = port
	return c
}
func (c *eurekaConfig) WithIp(ip string) *eurekaConfig {
	c.ip = ip
	return c
}
func (c *eurekaConfig) WithHostname(name string) *eurekaConfig {
	c.hostname = name
	return c
}

// TODO: WithMeta - поддержка множественных вызовов для установки разных данных

func (c *eurekaConfig) getInstanceId() string {
	if c.instanceId != "" {
		return c.instanceId
	}
	return fmt.Sprintf("%s:%s:%d", c.application, c.getIP(), c.port)
}
func (c *eurekaConfig) getIP() string {
	if c.ip != "" {
		return c.ip
	}
	return util.GetLocalIP()
}
func (c *eurekaConfig) getHostname() string {
	if c.hostname == "" {
		v, err := os.Hostname()
		if err != nil {
			logging.GetLogger("eureka-config").Warning("could not get hostname: %s", err)
			return "localhost"
		}
		return v
	}
	return c.hostname
}
