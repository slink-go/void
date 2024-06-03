package discovery

import (
	"encoding/json"
	"fmt"
	"github.com/slink-go/api-gateway/discovery/util"
	"github.com/slink-go/logging"
	"io"
	"os"
	"strings"
)

type Provider struct {
	logger logging.Logger
	config map[string][]Remote
}

func LoadFromFile(path string) (Client, error) {

	type registryConfigRecord struct {
		Name      string   `json:"name,omitempty",yaml:"name,omitempty"`
		Instances []string `json:"instances,omitempty",yaml:"instances,omitempty"`
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var cfg []registryConfigRecord
	if strings.HasSuffix(path, "yml") || strings.HasSuffix(path, "yaml") {
		// read yaml
		err = yaml.Unmarshal(bytes, &cfg)
	} else if strings.HasSuffix(path, "json") {
		// read json
		err = json.Unmarshal(bytes, &cfg)
	} else {
		err = fmt.Errorf("unsupported file type: %s", path)
	}
	if err != nil {
		return nil, err
	}

	result := Remotes{}
	for _, service := range cfg {
		for _, instance := range service.Instances {
			s, h, p := util.ParseEndpoint(instance)
			result.Add(service.Name, Remote{
				App:    service.Name,
				Scheme: s,
				Host:   h,
				Port:   p,
				Status: "UP",
			})
		}
	}
	return NewStaticClient(result.All()), nil
}
func NewStaticClient(services map[string][]Remote) Client {
	return &Provider{
		logger: logging.GetLogger("static-client"),
		config: services,
	}
}

func (c *Provider) Connect() error {
	return nil
}
func (c *Provider) Services() *Remotes {

	if c.config == nil {
		return nil
	}

	result := Remotes{}
	for app, remotes := range c.config {
		for _, instance := range remotes {
			r := Remote{
				App:    strings.ToUpper(app),
				Scheme: instance.Scheme,
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
