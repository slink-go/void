package gateway

import (
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
)

type Gateway interface {
	WithAuthProvider(ap security.AuthProvider) Gateway
	WithUserDetailsProvider(udp security.UserDetailsProvider) Gateway
	WithReverseProxy(reverseProxy *proxy.ReverseProxy) Gateway

	// WithRateLimiter(reverseProxy *proxy.ReverseProxy) Gateway

	Serve(address string)
}
