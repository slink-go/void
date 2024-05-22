package main

import (
	"github.com/slink-go/api-gateway/cmd/common"
)

func main() {

	common.LoadEnv()

	go Create("service-a", "A1", ":3101", "eureka")
	go Create("service-a", "A2", ":3102", "eureka")
	go Create("service-a", "A3", ":3103", "disco")
	go Create("service-a", "A4", ":3104", "disco")

	go Create("service-b", "B1", ":3201", "eureka")
	go Create("service-b", "B2", ":3202", "eureka")
	go Create("service-b", "B3", ":3203", "disco")
	go Create("service-b", "B4", ":3204", "disco")

	<-make(chan int)
}
