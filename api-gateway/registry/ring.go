package registry

import (
	"container/ring"
	"github.com/slink-go/api-gateway/discovery"
	"sync"
)

type ringBuffers struct {
	sync.RWMutex
	clientRing map[string]*ring.Ring
}

func createRingBuffers() *ringBuffers {
	return &ringBuffers{
		clientRing: make(map[string]*ring.Ring),
	}
}
func (b *ringBuffers) Get(serviceId string) (*ring.Ring, bool) {
	b.RLock()
	v, ok := b.clientRing[serviceId]
	b.RUnlock()
	return v, ok
}
func (b *ringBuffers) New(serviceId string, size int) {
	b.Lock()
	b.clientRing[serviceId] = ring.New(size)
	b.Unlock()
}
func (b *ringBuffers) Set(serviceId string, url *discovery.Remote) {
	b.Lock()
	b.clientRing[serviceId].Value = url
	b.clientRing[serviceId] = b.clientRing[serviceId].Next()
	b.Unlock()
}
func (b *ringBuffers) Next(serviceId string) (*ring.Ring, bool) {
	if b == nil || b.clientRing == nil {
		return nil, false
	}
	b.Lock()
	v, ok := b.clientRing[serviceId]
	if ok {
		b.clientRing[serviceId] = b.clientRing[serviceId].Next()
	}
	b.Unlock()
	return v, ok
}
func (b *ringBuffers) List() []any {
	if b == nil || b.clientRing == nil {
		return []any{}
	}
	b.RLock()
	defer b.RUnlock()
	result := make([]any, 0)
	for _, ring := range b.clientRing {
		for i := 0; i < ring.Len(); i++ {
			result = append(result, ring.Value)
			ring = ring.Move(1)
		}
	}

	return result
}
