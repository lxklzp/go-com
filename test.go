package main

import (
	"go-com/config"
	"go-com/core/logr"
	"go-com/internal/grpcc"
)

func main() {
	config.Load()
	logr.InitLog("app")

	grpcc.Client.Connect()
	grpcc.Client.Close()
}
