package resolver

import (
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/logging"
	"strings"
)

type ServiceResolver interface {
	Resolve(serviceName string) (string, error)
}

type serviceResolverImpl struct {
	serviceRegistry registry.ServiceRegistry
	logger          logging.Logger
}

func (sr *serviceResolverImpl) Resolve(serviceName string) (string, error) {
	result, err := sr.serviceRegistry.Get(strings.ToUpper(serviceName))
	if err != nil {
		sr.logger.Trace("%s -> %s ", serviceName, err)
	} else {
		sr.logger.Trace("%s -> %s ", serviceName, result)
	}
	return result, err
}

func NewServiceResolver(serviceRegistry registry.ServiceRegistry) ServiceResolver {
	return &serviceResolverImpl{
		serviceRegistry: serviceRegistry,
		logger:          logging.GetLogger("service-resolver"),
	}
}
