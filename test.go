package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/logr"
)

func main() {
	config.Load()
	logr.InitLog("test")

	fmt.Println(network.IPv4CheckWhitelist("195.168.3.1111", []string{
		"193.*.*.*",
		"192.*.*.*",
		"*.*",
		"195.168.3.64",
		"kk",
	}))
}
