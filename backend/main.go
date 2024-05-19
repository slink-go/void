package main

import (
	"os"
)

func main() {
	os.Setenv("GO_ENV", "dev")

	go Create("A1", ":3101")
	go Create("A2", ":3102")
	go Create("A3", ":3103")

	go Create("B1", ":3201")
	go Create("B2", ":3202")
	go Create("B3", ":3203")

	<-make(chan int)
}
