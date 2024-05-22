package disco

import (
	"fmt"
	"github.com/slink-go/api-gateway/cmd/common/util"
	"github.com/slink-go/api-gateway/discovery"
	d "github.com/slink-go/disco-go"
	da "github.com/slink-go/disco/common/api"
	"github.com/slink-go/logger"
	"github.com/slink-go/logging"
	"os"
	"strings"
	"sync"
	"time"
)

func NewDiscoClient(config *Config) discovery.Client {
	return &Client{
		config: *config,
		logger: logging.GetLogger("disco-client"),
		sigChn: make(chan os.Signal),
	}
}

type Client struct {
	config Config
	mutex  sync.RWMutex
	client d.DiscoClient
	logger logging.Logger
	sigChn chan os.Signal
}

func (c *Client) Connect() error {
	for {
		cfg := d.
			EmptyConfig().
			WithDisco([]string{c.config.url}).
			//WithBreaker(c.config.breakThreshold).
			WithRetry(c.config.retryAttempts, c.config.retryDelay).
			WithTimeout(c.config.timeout).
			WithAuth(c.config.login, c.config.password).
			WithName(c.config.application).
			WithEndpoints([]string{fmt.Sprintf("%s://%s:%d", "http", c.config.getIP(), c.config.port)})
		//WithMeta()
		clnt, err := d.NewDiscoHttpClient(cfg)
		if err != nil {
			logger.Warning("join error: %s", strings.TrimSpace(err.Error()))
			time.Sleep(5 * time.Second) // TODO: need configurable retry interval
			continue
		}
		c.client = clnt
		break
	}
	return nil
}
func (c *Client) Services() *discovery.Remotes {
	if c.client == nil {
		return nil
	}
	result := discovery.Remotes{}
	appId := strings.ToUpper(c.config.application)
	for _, v := range c.client.Registry().List() {
		if da.ClientStateUp == v.State() && v.ServiceId() != appId {
			ep, err := v.Endpoint(da.HttpEndpoint)
			if err != nil {
				c.logger.Warning("could not find HTTP endpoint for %s", v.ServiceId())
				continue
			}
			host, port := util.ParseEndpoint(ep)
			remote := discovery.Remote{
				App:    v.ServiceId(),
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
