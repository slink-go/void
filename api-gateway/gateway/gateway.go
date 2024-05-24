package gateway

import (
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
	"github.com/slink-go/api-gateway/registry"
)

type Gateway interface {
	WithAuthProvider(ap security.AuthProvider) Gateway
	WithUserDetailsProvider(udp security.UserDetailsProvider) Gateway
	WithRateLimiter(limiter rate.Limiter) Gateway
	WithReverseProxy(reverseProxy *proxy.ReverseProxy) Gateway
	WithRegistry(registry registry.ServiceRegistry) Gateway
	WithQuitChn(chn chan struct{}) Gateway
	Serve(addresses ...string)
}
