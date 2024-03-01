package main

import (
	"context"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/api"
	"os"
	"strconv"
)

func main() {
	config.Load()
	logr.InitLog("main")
	config.InitDefine()
	logr.L.Debug("启动系统:" + strconv.Itoa(os.Getpid()))

	api.Run()

	tool.ExitNotify(func() {
		api.Shutdown()
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
