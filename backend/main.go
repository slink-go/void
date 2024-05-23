package main

import (
	"github.com/slink-go/api-gateway/cmd/common"
	"github.com/slink-go/logging"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	common.LoadEnv()

	services := []*Service{
		Create("service-a", "A1", ":3101", "eureka"),
		Create("service-a", "A2", ":3102", "eureka"),
		Create("service-b", "B1", ":3201", "eureka"),
		Create("service-b", "B2", ":3202", "eureka"),
		Create("service-a", "A3", ":3103", "disco"),
		Create("service-a", "A4", ":3104", "disco"),
		Create("service-b", "B3", ":3203", "disco"),
		Create("service-b", "B4", ":3204", "disco"),
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
		<-make(chan int)
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
			time.Sleep(time.Second)
			os.Exit(0)
		}
	}
}
