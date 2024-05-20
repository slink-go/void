package discovery

import (
	"github.com/slink-go/api-gateway/registry"
)

type Client interface {
	Register() error
	Get(string) (string, error)
	Services() map[string][]registry.Remote
}
