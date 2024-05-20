package gateway

import (
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
)

type Gateway interface {
	WithAuthProvider(ap security.AuthProvider) Gateway
	WithUserDetailsProvider(udp security.UserDetailsProvider) Gateway
	WithRateLimiter(limiter rate.Limiter) Gateway
	WithReverseProxy(reverseProxy *proxy.ReverseProxy) Gateway
	Serve(address string)
}
