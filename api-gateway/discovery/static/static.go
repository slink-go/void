package static

import (
	"github.com/slink-go/api-gateway/discovery"
	"github.com/slink-go/logging"
	"strings"
)

type Provider struct {
	logger logging.Logger
	config map[string][]discovery.Remote
}

func NewStaticClient(services map[string][]discovery.Remote) discovery.Client {
	return &Provider{
		logger: logging.GetLogger("static-client"),
		config: services,
	}
}

func (c *Provider) Connect() error {
	return nil
}
func (c *Provider) Services() *discovery.Remotes {

	if c.config == nil {
		return nil
	}

	result := discovery.Remotes{}
	for app, remotes := range c.config {
		for _, instance := range remotes {
			r := discovery.Remote{
				App:    strings.ToUpper(app),
				Host:   instance.Host,
				Port:   instance.Port,
				Status: "UP",
			}
			r.Status = instance.Status
			c.logger.Debug("add %s: %s", instance.App, r)
			result.Add(app, r)
		}
	}

	return &result

}
