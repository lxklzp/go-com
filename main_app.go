package main

import (
	"context"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/api"
	"go-com/internal/app"
	"go-com/internal/system"
	"os"
	"strconv"
)

func main() {
	config.Load()
	logr.InitLog("app")
	system.RunApp()

	logr.L.Info("启动系统:" + strconv.Itoa(os.Getpid()))
	tool.ExitNotify(func() {
		api.Shutdown()
		app.Cron.Stop()
	})

	// 启动接口服务
	api.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
