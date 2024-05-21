package eureka

import (
	"errors"
	"fmt"
	"github.com/slink-go/api-gateway/discovery"
	e "github.com/slink-go/go-eureka-client/eureka"
	"github.com/slink-go/logging"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

func NewEurekaClient(config *Config) discovery.Client {

	return &EurekaClient{
		config: *config,
		logger: logging.GetLogger("eureka-client"),
		sigChn: make(chan os.Signal),
	}
}

type EurekaClient struct {
	config       Config
	mutex        sync.RWMutex
	client       *e.Client
	logger       logging.Logger
	running      bool
	applications *e.Applications
	sigChn       chan os.Signal
}

func (c *EurekaClient) Connect() error {

	c.mutex.Lock()
	c.running = true
	c.mutex.Unlock()

	if c.config.url == "" {
		return fmt.Errorf("eureka url is empty")
	}
	c.client = e.NewClient([]string{c.config.url})

	if c.config.login != "" && c.config.password != "" {
		c.client.WithBasicAuth(c.config.login, c.config.password)
	}

	if c.config.fetch {
		go c.refresh()
	}
	if c.config.register {
		go c.heartbeat()
	}

	go c.handleSignal()

	return nil
}
func (c *EurekaClient) Services() *discovery.Remotes {

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if c.applications == nil {
		return nil
	}

	result := discovery.Remotes{}
	for _, app := range c.applications.Applications {
		for _, instance := range app.Instances {
			r := discovery.Remote{
				App:  app.Name,
				Host: instance.IpAddr,
			}
			if instance.Port != nil {
				r.Port = instance.Port.Port
			}
			r.Status = instance.Status
			result.Add(app.Name, r)
		}
	}

	return &result

}

func (c *EurekaClient) create() {
}
func (c *EurekaClient) register() error {
	return c.client.RegisterInstance(c.config.application, c.createInstance())
}
func (c *EurekaClient) unregister() error {
	return c.client.UnregisterInstance(c.config.application, c.config.getInstanceId())
}
func (c *EurekaClient) refresh() {
	timer := time.NewTimer(time.Second)
	for c.running {
		select {
		case <-timer.C:
			apps, err := c.client.GetApplications()
			if err != nil {
				c.logger.Error("refresh failed: %s", err)
			} else {
				c.mutex.Lock()
				c.applications = apps
				c.mutex.Unlock()
				c.logger.Trace("[%s] refresh complete", c.config.getInstanceId())
			}
		}
		timer.Reset(c.config.refreshInterval)
	}
	timer.Stop()
}
func (c *EurekaClient) heartbeat() {
	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-timer.C:
			err := c.register()
			if err == nil {
				break
			}
			c.logger.Warning("registration failed: %s", err)
		}
		timer.Reset(c.config.heartBeatInterval)
	}
	for c.running {
		select {
		case <-timer.C:
			err := c.client.SendHeartbeat(c.config.application, c.config.getInstanceId())

			if err == nil {
				c.logger.Debug("heartbeat application instance successful")
			} else if err == ErrNotFound {
				// heartbeat not found, need register
				err = c.register()
				if err == nil {
					c.logger.Info("register application instance successful")
				} else {
					c.logger.Error("register application instance failed: %s", err)
				}
			} else {
				c.logger.Error("heartbeat application instance failed: %s", err)
			}
		}
		timer.Reset(c.config.heartBeatInterval)
	}
	timer.Stop()
}
func (c *EurekaClient) createInstance() *e.InstanceInfo {
	dcInfo := &e.DataCenterInfo{
		Name:  "MyOwn",
		Class: "com.netflix.appinfo.MyDataCenterInfo", //"com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
	}
	return &e.InstanceInfo{
		App:                           c.config.application,
		IpAddr:                        c.config.getIP(),
		VipAddress:                    c.config.application,
		SecureVipAddress:              c.config.application,
		Status:                        "UP",
		Port:                          &e.Port{Port: c.config.port, Enabled: true},
		DataCenterInfo:                dcInfo,
		IsCoordinatingDiscoveryServer: false,
		LastUpdatedTimestamp:          0,
		LastDirtyTimestamp:            0,
		ActionType:                    "",
		Overriddenstatus:              "UNKNOWN",
		CountryId:                     0,
		InstanceID:                    c.config.getInstanceId(),
		HomePageUrl:                   fmt.Sprintf("http://%s:%d", c.config.getIP(), c.config.port),
		HostName:                      c.config.hostname,
		//StatusPageUrl:  "", // fmt.Sprintf("http://%s:%d/info", config.IP, config.Port)
		//HealthCheckUrl: "",
		//SecurePort:       "",
		//LeaseInfo: &e.LeaseInfo{RenewalIntervalInSecs: c.config.leaseRenewalInterval, DurationInSecs: c.config.leaseDurationInSecs},
		//Metadata:                      nil, // c.config.metadata
	}
	//return e.NewInstanceInfo(
	//	c.config.hostname,
	//	c.config.application,
	//	c.config.getIP(),
	//	c.config.port,
	//	30,
	//	false,
	//)
}
func (c *EurekaClient) repeat(interval time.Duration, action func(), stopChn <-chan struct{}) {
}
func (c *EurekaClient) handleSignal() {
	signal.Notify(c.sigChn, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		switch <-c.sigChn {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGTERM:
			c.logger.Info("receive exit signal")
			if c.config.register {
				c.logger.Info("client instance going to de-register")
				err := c.unregister()
				if err != nil {
					c.logger.Error("application instance de-registration failed: %s", err)
				} else {
					c.logger.Info("application instance de-registered")
				}
			}
			os.Exit(0)
		}
	}
}
