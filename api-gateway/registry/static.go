package registry

func NewStaticRegistry(services map[string][]Remote) ServiceRegistry {
	directory := createRingBuffers()
	for k, list := range services {
		directory.New(k, len(list))
		for _, url := range list {
			v := url
			directory.Set(k, &v)
		}
	}
	return &staticRegistry{
		serviceDirectory: directory,
	}
}

type staticRegistry struct {
	serviceDirectory *ringBuffers
}

func (sr *staticRegistry) Get(serviceName string) (string, error) {
	v, ok := sr.serviceDirectory.Next(serviceName)
	if !ok {
		return "", NewErrServiceUnavailable(serviceName)
	}
	url := v.Next().Value.(*Remote)
	return (*url).String(), nil
}
