package discovery

import (
	"fmt"
	"github.com/slink-go/api-gateway/discovery/util"
	d "github.com/slink-go/disco-go"
	da "github.com/slink-go/disco/common/api"
	"github.com/slink-go/logging"
	"os"
	"strings"
	"sync"
	"time"
)

func NewDiscoClient(config *discoConfig) Client {
	return &discoClient{
		config:        *config,
		logger:        logging.GetLogger("disco-client"),
		sigChn:        make(chan os.Signal),
		Notifications: nil,
	}
}

type discoClient struct {
	config        discoConfig
	mutex         sync.RWMutex
	client        d.DiscoClient
	logger        logging.Logger
	sigChn        chan os.Signal
	Notifications chan struct{}
}

func (c *discoClient) Connect(options ...interface{}) error {
	if c.config.url == "" {
		return fmt.Errorf("disco url is empty")
	}
	cfg := d.
		EmptyConfig().
		WithDisco([]string{c.config.url}).
		//WithBreaker(c.config.breakThreshold).
		WithRetry(c.config.retryAttempts, c.config.retryDelay).
		WithTimeout(c.config.timeout).
		WithAuth(c.config.login, c.config.password).
		WithName(c.config.application).
		WithEndpoints([]string{fmt.Sprintf("%s://%s:%d", "http", c.config.getIP(), c.config.port)})

	if len(options) > 0 {
		if chn, ok := options[0].(chan struct{}); ok {
			cfg.WithNotificationChn(chn)
			c.Notifications = chn
		}
	}

	//WithMeta()
	go func() {
		for {
			clnt, err := d.NewDiscoHttpClient(cfg)
			if err != nil {
				c.logger.Warning("join error: %s", strings.TrimSpace(err.Error()))
				time.Sleep(5 * time.Second) // TODO: need configurable retry interval
				continue
			}
			c.client = clnt
			break
		}
	}()
	return nil
}
func (c *discoClient) Services() *Remotes {
	if c.client == nil {
		return nil
	}
	result := Remotes{}
	appId := strings.ToUpper(c.config.application)
	for _, v := range c.client.Registry().List() {
		if da.ClientStateUp == v.State() && v.ServiceId() != appId {
			ep, err := v.Endpoint(da.HttpEndpoint)
			if err != nil {
				c.logger.Warning("could not find HTTP endpoint for %s", v.ServiceId())
				continue
			}
			scheme, host, port := util.ParseEndpoint(ep)
			remote := Remote{
				App:    v.ServiceId(),
				Scheme: scheme,
				Host:   host,
				Port:   port,
				Status: "UP",
			}
			c.logger.Debug("add %s: %s", v.ServiceId(), remote)
			result.Add(v.ServiceId(), remote)
		}
	}
	return &result
}
func (c *discoClient) NotificationsChn() chan struct{} {
	return c.Notifications
}
