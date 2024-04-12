package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/logr"
)

func main() {
	config.Load()
	logr.InitLog("test")

	fmt.Println(3.0 / 7.0)
}
