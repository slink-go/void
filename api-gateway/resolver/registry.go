package resolver

import (
	"container/ring"
	"fmt"
)

type ServiceRegistry interface {
	Get(serviceName string) (string, error)
}

type staticServiceRegistry struct {
	serviceDirectory map[string]ring.Ring
}

func (sr *staticServiceRegistry) Get(serviceName string) (string, error) {
	v, ok := sr.serviceDirectory[serviceName]
	if !ok {
		return "", NewErrServiceUnavailable(serviceName)
	}
	return fmt.Sprintf("%s", v.Next().Value), nil
}

func NewStaticServiceRegistry(serviceDirectory map[string][]string) ServiceRegistry {
	directory := make(map[string]ring.Ring)
	for k, list := range serviceDirectory {
		rng := ring.New(len(list))
		for _, url := range list {
			rng.Value = url
			rng.Next()
		}
		directory[k] = *rng
	}
	return &staticServiceRegistry{
		serviceDirectory: directory,
	}
}
