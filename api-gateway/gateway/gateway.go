package gateway

import (
	"github.com/slink-go/api-gateway/middleware/rate"
	"github.com/slink-go/api-gateway/middleware/security"
	"github.com/slink-go/api-gateway/proxy"
)

type Gateway interface {
	WithAuthProvider(ap security.AuthProvider) Gateway
	WithUserDetailsProvider(udp security.UserDetailsProvider) Gateway
	WithReverseProxy(reverseProxy *proxy.ReverseProxy) Gateway
	WithRateLimiter(limiter rate.Limiter) Gateway
	Serve(address string)
}
