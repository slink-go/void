package env

const (
	GoEnv                   = "GO_ENV"
	ServicePort             = "SERVICE_PORT"              // service port ( + port for registration in eureka)
	EurekaUrl               = "EUREKA_URL"                //
	EurekaLogin             = "EUREKA_LOGIN"              //
	EurekaPassword          = "EUREKA_PASSWORD"           //
	EurekaHeartbeatInterval = "EUREKA_HEARTBEAT_INTERVAL" // default 10s
	EurekaRefreshInterval   = "EUREKA_REFRESH_INTERVAL"   // default 30s
	RegistryRefreshInterval = "REGISTRY_REFRESH_INTERVAL" // default 60s
)
