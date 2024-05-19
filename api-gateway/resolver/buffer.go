package resolver

import (
	"container/ring"
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
func (b *ringBuffers) Set(serviceId string, url *string) {
	b.Lock()
	b.clientRing[serviceId].Value = url
	b.clientRing[serviceId] = b.clientRing[serviceId].Next()
	b.Unlock()
}
func (b *ringBuffers) Next(serviceId string) (*ring.Ring, bool) {
	b.Lock()
	v, ok := b.clientRing[serviceId]
	b.clientRing[serviceId] = b.clientRing[serviceId].Next()
	b.Unlock()
	return v, ok
}
