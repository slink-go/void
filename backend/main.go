package main

import (
	"os"
)

func main() {
	os.Setenv("GO_ENV", "dev")

	go Create("SERVICE-A", "A1", ":3101")
	go Create("SERVICE-A", "A2", ":3102")
	go Create("SERVICE-A", "A3", ":3103")

	go Create("SERVICE-B", "B1", ":3201")
	go Create("SERVICE-B", "B2", ":3202")
	go Create("SERVICE-B", "B3", ":3203")

	<-make(chan int)
}
