package main

import (
	"fmt"
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/api-gateway/cmd/common/env"
	"github.com/slink-go/logging"
	"os"
	"os/signal"
	"syscall"
)

func getPort(base, port int) string {
	return fmt.Sprintf(":%d", base+port)
}
func main() {

	common.LoadEnv()
	basePort := int(env.Int64OrDefault("BASE_PORT", 3000))

	services := []*Service{
		Create("service-a", "A1", getPort(basePort, 101), "eureka"),
		Create("service-a", "A2", getPort(basePort, 102), "eureka"),
		Create("service-b", "B1", getPort(basePort, 201), "eureka"),
		Create("service-b", "B2", getPort(basePort, 202), "eureka"),
		Create("service-a", "A3", getPort(basePort, 103), "disco"),
		Create("service-a", "A4", getPort(basePort, 104), "disco"),
		Create("service-b", "B3", getPort(basePort, 203), "disco"),
		Create("service-b", "B4", getPort(basePort, 204), "disco"),
	}

	started := 0
	for _, service := range services {
		if service != nil {
			logging.GetLogger("main").Info("start at %s", service.address)
			go service.Start()
			started++
		}
	}

	if started == 0 {
		logging.GetLogger("main").Info("no services started, exiting")

	} else {
		logging.GetLogger("main").Info("running...")
		handleSignal()
	}

}

func handleSignal() {
	sigChn := make(chan os.Signal, 1)
	signal.Notify(sigChn, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
	for {
		switch <-sigChn {
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGKILL:
			fallthrough
		case syscall.SIGTERM:
			os.Exit(0)
		}
	}
}
