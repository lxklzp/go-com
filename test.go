package main

import (
	"context"
	"fmt"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/rds"
	"go-com/internal/app"
)

func main() {
	config.Load()
	logr.InitLog("test")
	app.Redis = rds.NewRedis(rds.Config{Redis: config.C.Redis})
	fmt.Println(app.Redis.Info(context.TODO()).String())
}
