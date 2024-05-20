package registry

import (
	"github.com/slink-go/api-gateway/discovery"
	"sync"
)

type discoveryRegistry struct {
	serviceDirectory *ringBuffers // TODO: implement periodic refresh | при обновлении собьётся ringBuffer; можно ли что-то с этим сделать? стоит ли это делать?
	client           discovery.Client
	mutex            sync.RWMutex
}

func NewDiscoveryRegistry(client discovery.Client) ServiceRegistry {
	registry := discoveryRegistry{
		serviceDirectory: nil,
		client:           client,
	}
	registry.refresh()
	return &registry
}

func (sr *discoveryRegistry) refresh() {
	directory := createRingBuffers()
	for k, list := range sr.client.Services() {
		directory.New(k, len(list))
		for _, url := range list {
			v := url
			directory.Set(k, &v)
		}
	}
	sr.mutex.Lock()
	sr.serviceDirectory = directory
	sr.mutex.Unlock()
}

func (sr *discoveryRegistry) Get(serviceName string) (string, error) {
	sr.mutex.RLock()
	defer sr.mutex.RUnlock()
	v, ok := sr.serviceDirectory.Next(serviceName)
	if !ok {
		return "", NewErrServiceUnavailable(serviceName)
	}
	url := v.Next().Value.(*discovery.Remote)
	return (*url).String(), nil
}
