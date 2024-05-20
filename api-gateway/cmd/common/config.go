package common

import (
	"github.com/joho/godotenv"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/api-gateway/registry"
	"github.com/slink-go/logging"
	"os"
)

func Services() map[string][]registry.Remote {
	var services = make(map[string][]registry.Remote, 2)
	services["service-a"] = []registry.Remote{
		registry.Remote{
			Port: 3101,
			Host: "backend",
		},
		registry.Remote{
			Port: 3102,
			Host: "backend",
		},
		registry.Remote{
			Port: 3103,
			Host: "backend",
		},
	}
	services["service-b"] = []registry.Remote{
		registry.Remote{
			Port: 3201,
			Host: "backend",
		},
		registry.Remote{
			Port: 3202,
			Host: "backend",
		},
		registry.Remote{
			Port: 3203,
			Host: "backend",
		},
	}
	return services
}
func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		os.Setenv(env.GoEnv, "dev")
		logging.GetLogger("main").Warning("could not read config from .env file")
	}
}
func GetServicePorts() (base int, add int) {
	base = int(env.Int64OrDefault(env.ServicePort, 0))
	if base > 0 {
		add = base + 1
	}
	if base <= 0 && add <= 0 {
		panic("service ports not set")
	}
	return
}
