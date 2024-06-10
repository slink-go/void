package common

import (
	"flag"
	"github.com/joho/godotenv"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/logging"
	"os"
)

func LoadEnv() {
	configFilePathPtr := flag.String("config", ".env", "configuration file path")
	flag.Parse()

	var err error
	if configFilePathPtr != nil {
		err = godotenv.Load(*configFilePathPtr)
	} else {
		err = godotenv.Load(".env")
		logging.GetLogger("main").Info("read environment config")
	}
	if err != nil {
		os.Setenv(env.GoEnv, "dev")
		logging.GetLogger("main").Warning("could not read config from .env file")
	}
}
