package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/rds"
	"go-com/core/security"
	"go-com/internal/app"
)

func main() {
	config.Load()
	logr.InitLog("test")
	app.Redis = rds.NewRedis(rds.Config{Redis: config.C.Redis})
	fmt.Println(security.Md5Encrypt("aa"))
}
