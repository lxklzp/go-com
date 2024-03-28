package main

import (
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/my"
	"go-com/internal/app"
)

func main() {
	config.Load()
	logr.InitLog("test")
	app.Mysql = my.NewDb(my.Config{Mysql: config.C.Mysql})

}
