package main

import (
	"context"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/system"
	"go-com/internal/webapi"
	"os"
	"strconv"
)

func main() {
	config.Load()
	logr.InitLog("web")
	system.RunWeb()
	logr.L.Info("启动系统:" + strconv.Itoa(os.Getpid()))
	tool.ExitNotify(func() {
		webapi.Shutdown()
	})

	// 启动接口服务
	webapi.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
