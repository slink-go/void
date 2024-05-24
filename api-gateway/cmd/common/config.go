package common

import (
	"github.com/joho/godotenv"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/logging"
	"os"
)

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		os.Setenv(env.GoEnv, "dev")
		logging.GetLogger("main").Warning("could not read config from .env file")
	}
}
