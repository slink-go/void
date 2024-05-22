package env

const (
	GoEnv       = "GO_ENV"
	ServicePort = "SERVICE_PORT" // service port ( + port for registration in eureka)

	EurekaUrl               = "EUREKA_URL"                //
	EurekaLogin             = "EUREKA_LOGIN"              //
	EurekaPassword          = "EUREKA_PASSWORD"           //
	EurekaHeartbeatInterval = "EUREKA_HEARTBEAT_INTERVAL" // default 10s
	EurekaRefreshInterval   = "EUREKA_REFRESH_INTERVAL"   // default 30s

	DiscoUrl      = "DISCO_URL"
	DiscoLogin    = "DISCO_LOGIN"
	DiscoPassword = "DISCO_PASSWORD"

	RegistryRefreshInterval = "REGISTRY_REFRESH_INTERVAL" // default 60s
	RateLimitRPM            = "RATE_LIMIT_RPM"            // rate limit "requests per minute" (for fiber limiter)
	RateLimitRPS            = "RATE_LIMIT_RPS"            // rate limit "requests per second" (for gin limiter)
)
