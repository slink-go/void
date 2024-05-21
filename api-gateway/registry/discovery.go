package registry

import (
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/discovery"
	"github.com/slink-go/logging"
	"strings"
	"sync"
	"time"
)

type discoveryRegistry struct {
	serviceDirectory *ringBuffers // TODO: implement periodic refresh | при обновлении собьётся ringBuffer; можно ли что-то с этим сделать? стоит ли это делать?
	client           discovery.Client
	mutex            sync.RWMutex
	logger           logging.Logger
}

func NewDiscoveryRegistry(client discovery.Client) ServiceRegistry {
	registry := discoveryRegistry{
		serviceDirectory: nil,
		client:           client,
		logger:           logging.GetLogger("discovery-registry"),
	}
	go registry.refresh()
	return &registry
}

func (sr *discoveryRegistry) refresh() {
	timer := time.NewTimer(5 * time.Second)
	interval := env.DurationOrDefault(env.RegistryRefreshInterval, time.Second*60)
	for {
		select {
		case <-timer.C:
			remotes := make(map[string][]discovery.Remote)
			directory := createRingBuffers()
			for _, instance := range sr.client.Services().List() {
				appId := strings.ToUpper(instance.App)
				if _, ok := remotes[appId]; !ok {
					remotes[appId] = make([]discovery.Remote, 0)
				}
				sr.logger.Trace("add item: %T %s", instance, instance)
				remotes[appId] = append(remotes[appId], instance)
			}
			for k, list := range remotes {
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
		timer.Reset(interval)
	}
	timer.Stop()
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
func (sr *discoveryRegistry) List() []discovery.Remote {
	result := make([]discovery.Remote, 0)
	for _, v := range sr.serviceDirectory.List() {
		vv := v.(*discovery.Remote)
		sr.logger.Trace("list item: %s", vv)
		result = append(result, *vv)
	}
	return result
}
