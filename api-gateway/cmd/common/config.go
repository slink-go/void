package common

import (
	"github.com/joho/godotenv"
	"github.com/slink-go/api-gateway/resolver"
	"github.com/slink-go/logging"
	"os"
)

func serviceRegistry() resolver.ServiceRegistry {
	var registry = make(map[string][]string, 2)
	registry["service-a"] = []string{"backend:3101", "backend:3102", "backend:3103"}
	registry["service-b"] = []string{"backend:3201", "backend:3202", "backend:3203"}
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
