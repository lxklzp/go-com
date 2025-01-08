package main

import (
	"go-com/config"
	"go-com/core/logr"
)

func main() {
	config.Load()
	logr.InitLog("test")

}
