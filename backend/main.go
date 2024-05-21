package main

import (
	"github.com/slink-go/api-gateway/cmd/common"
)

func main() {

	common.LoadEnv()

	go Create("service-a", "A1", ":3101")
	go Create("service-a", "A2", ":3102")
	go Create("service-a", "A3", ":3103")

	go Create("service-b", "B1", ":3201")
	go Create("service-b", "B2", ":3202")
	go Create("service-b", "B3", ":3203")

	<-make(chan int)
}
