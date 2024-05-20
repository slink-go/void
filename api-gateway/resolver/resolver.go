package resolver

import "github.com/slink-go/api-gateway/registry"

type ServiceResolver interface {
	Resolve(serviceName string) (string, error)
}

type serviceResolverImpl struct {
	serviceRegistry registry.ServiceRegistry
}

func (sr *serviceResolverImpl) Resolve(serviceName string) (string, error) {
	return sr.serviceRegistry.Get(serviceName)
}

func NewServiceResolver(serviceRegistry registry.ServiceRegistry) ServiceResolver {
	return &serviceResolverImpl{
		serviceRegistry: serviceRegistry,
	}
}
