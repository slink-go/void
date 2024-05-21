package registry

import "github.com/slink-go/api-gateway/discovery"

func NewStaticRegistry(services map[string][]discovery.Remote) ServiceRegistry {
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
	url := v.Next().Value.(*discovery.Remote)
	return (*url).String(), nil
}
func (sr *staticRegistry) List() []discovery.Remote {
	result := make([]discovery.Remote, 0)
	for _, v := range sr.serviceDirectory.List() {
		result = append(result, *v.(*discovery.Remote))
	}
	return result
}
