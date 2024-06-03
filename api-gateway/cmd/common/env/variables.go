package env

const (
	GoEnv = "GO_ENV"

	GatewayName       = "GATEWAY_NAME"
	ServicePort       = "SERVICE_PORT" // service port ( + port for registration in eureka)
	MonitoringEnabled = "MONITORING_ENABLED"
	MonitoringPort    = "MONITORING_PORT"

	AuthEnabled                 = "AUTH_ENABLED"
	AuthEndpoint                = "AUTH_ENDPOINT"
	AuthMethod                  = "AUTH_METHOD"
	AuthResponseMappingFilePath = "AUTH_RESPONSE_MAPPING_FILE_PATH"
	AuthCacheTTL                = "AUTH_CACHE_TTL"

	EurekaClientEnabled     = "EUREKA_CLIENT_ENABLED"
	EurekaUrl               = "EUREKA_URL"                //
	EurekaLogin             = "EUREKA_LOGIN"              //
	EurekaPassword          = "EUREKA_PASSWORD"           //
	EurekaHeartbeatInterval = "EUREKA_HEARTBEAT_INTERVAL" // default 10s
	EurekaRefreshInterval   = "EUREKA_REFRESH_INTERVAL"   // default 30s

	DiscoClientEnabled = "DISCO_CLIENT_ENABLED"
	DiscoUrl           = "DISCO_URL"
	DiscoLogin         = "DISCO_LOGIN"
	DiscoPassword      = "DISCO_PASSWORD"

	StaticRegistryFile = "STATIC_REGISTRY_FILE"

	RegistryRefreshInitialDelay = "REGISTRY_REFRESH_INITIAL_DELAY"
	RegistryRefreshInterval     = "REGISTRY_REFRESH_INTERVAL" // default 60s
	RateLimitRPM                = "RATE_LIMIT_RPM"            // rate limit "requests per minute" (for fiber limiter)
)
