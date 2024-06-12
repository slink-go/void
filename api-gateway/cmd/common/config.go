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
		//logging.GetLogger("main").Info("try load environment config from %s", *configFilePathPtr)
		err = godotenv.Load(*configFilePathPtr)
	} else {
		//logging.GetLogger("main").Info("try load environment config from .env")
		err = godotenv.Load(".env")
	}
	if err != nil {
		os.Setenv(env.GoEnv, "dev")
		logging.GetLogger("main").Warning("could not read config from %s", *configFilePathPtr)
	} else {
		logging.GetLogger("main").Info("environment config loaded from %s", *configFilePathPtr)
	}

}
