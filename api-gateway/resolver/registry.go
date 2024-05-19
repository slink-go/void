package resolver

type ServiceRegistry interface {
	Get(serviceName string) (string, error)
}

type staticServiceRegistry struct {
	serviceDirectory *ringBuffers
}

func (sr *staticServiceRegistry) Get(serviceName string) (string, error) {
	v, ok := sr.serviceDirectory.Next(serviceName)
	if !ok {
		return "", NewErrServiceUnavailable(serviceName)
	}
	url := v.Next().Value.(*string)
	return *url, nil
}

func NewStaticServiceRegistry(serviceDirectory map[string][]string) ServiceRegistry {
	directory := createRingBuffers()
	for k, list := range serviceDirectory {
		directory.New(k, len(list))
		for _, url := range list {
			v := url
			directory.Set(k, &v)
		}
	}
	return &staticServiceRegistry{
		serviceDirectory: directory,
	}
}
