package main

import (
	"fmt"

	"./config"
)

func main() {
	fmt.Println("starting scanner")
	err := config.ConfigureInterfaces()
	fmt.Println(err)
}
