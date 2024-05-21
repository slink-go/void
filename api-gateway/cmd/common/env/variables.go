package env

const (
	GoEnv                   = "GO_ENV"
	ServicePort             = "SERVICE_PORT"              // service port ( + port for registration in eureka)
	EurekaHeartbeatInterval = "HEARTBEAT_INTERVAL"        // default 10s
	EurekaRefreshInterval   = "REFRESH_INTERVAL"          // default 30s
	EurekaUrl               = "EUREKA_URL"                //
	EurekaLogin             = "EUREKA_LOGIN"              //
	EurekaPassword          = "EUREKA_PASSWORD"           //
	RegistryRefreshInterval = "REGISTRY_REFRESH_INTERVAL" // default 60s
)
