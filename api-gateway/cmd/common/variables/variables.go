package variables

const (
	GoEnv = "GO_ENV"

	GatewayName       = "GATEWAY_NAME"
	ServicePort       = "SERVICE_PORT" // service port ( + port for registration in eureka)
	MonitoringEnabled = "MONITORING_ENABLED"
	MonitoringPort    = "MONITORING_PORT"

	TargetConnTimeout         = "TARGET_CONN_TIMEOUT"
	TargetConnKeepAlive       = "TARGET_CONN_KEEPALIVE"
	TargetTLSHandshakeTimeout = "TARGET_TLS_HANDSHAKE_TIMEOUT"

	AuthEnabled                 = "AUTH_ENABLED"
	AuthEndpoint                = "AUTH_ENDPOINT"
	AuthMethod                  = "AUTH_METHOD"
	AuthResponseMappingFilePath = "AUTH_RESPONSE_MAPPING_FILE_PATH"
	AuthSkip                    = "AUTH_SKIP"
	AuthCacheTTL                = "AUTH_CACHE_TTL"

	RequestTimeout = "REQUEST_TIMEOUT"
	TimeoutSkip    = "TIMEOUT_SKIP"

	EurekaClientEnabled     = "EUREKA_CLIENT_ENABLED"
	EurekaUrl               = "EUREKA_URL"                //
	EurekaLogin             = "EUREKA_LOGIN"              //
	EurekaPassword          = "EUREKA_PASSWORD"           //
	EurekaHeartbeatInterval = "EUREKA_HEARTBEAT_INTERVAL" // default 10s
	EurekaRefreshInterval   = "EUREKA_REFRESH_INTERVAL"   // default 30s

	DiscoClientEnabled       = "DISCO_CLIENT_ENABLED"
	DiscoUrl                 = "DISCO_URL"
	DiscoLogin               = "DISCO_LOGIN"
	DiscoPassword            = "DISCO_PASSWORD"
	DiscoClientRetryInterval = "DISCO_CLIENT_RETRY_INTERVAL"

	StaticRegistryFile = "STATIC_REGISTRY_FILE"

	RegistryRefreshInitialDelay = "REGISTRY_REFRESH_INITIAL_DELAY"
	RegistryRefreshInterval     = "REGISTRY_REFRESH_INTERVAL" // default 60s

	LimiterLimit                  = "LIMITER_LIMIT"
	LimiterPeriod                 = "LIMITER_PERIOD"
	LimiterMode                   = "LIMITER_MODE"
	LimiterCustomConfig           = "LIMITER_CUSTOM"
	RateLimitCacheCleanupInterval = "RATE_LIMIT_CACHE_CLEANUP_INTERVAL"
)
