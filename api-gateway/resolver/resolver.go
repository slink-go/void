package resolver

type ServiceResolver interface {
	Resolve(serviceName string) (string, error)
}

type serviceResolverImpl struct {
	serviceRegistry ServiceRegistry
}

func (sr *serviceResolverImpl) Resolve(serviceName string) (string, error) {
	return sr.serviceRegistry.Get(serviceName)
}

func NewServiceResolver(serviceRegistry ServiceRegistry) ServiceResolver {
	return &serviceResolverImpl{
		serviceRegistry: serviceRegistry,
	}
}
