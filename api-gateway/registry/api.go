package registry

import (
	"github.com/slink-go/api-gateway/discovery"
)

type ServiceRegistry interface {
	Get(applicationId string) (string, error)
	List() []discovery.Remote
}
