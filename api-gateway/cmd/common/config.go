package common

import (
	"github.com/joho/godotenv"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"os"
)

func serviceRegistry() resolver.ServiceRegistry {
	var registry = make(map[string][]string, 2)
	registry["service-a"] = []string{"localhost:3101", "localhost:3102", "localhost:3103"}
	registry["service-b"] = []string{"localhost:3201", "localhost:3202", "localhost:3203"}
	return resolver.NewStaticServiceRegistry(registry)
}
func ServiceResolver() resolver.ServiceResolver {
	return resolver.NewServiceResolver(serviceRegistry())
}
func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		os.Setenv("GO_ENV", "dev")
		logging.GetLogger("main").Warning("could not read config from .env file")
	}
}
