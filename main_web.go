package main

import "C"
import (
	"context"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/pg"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/webapi"
	"os"
	"strconv"
)

func main() {
	config.Load()
	logr.InitLog("web")
	app.Pg = pg.NewDb(pg.Config{Postgresql: config.C.Postgresql})
	logr.L.Info("启动系统:" + strconv.Itoa(os.Getpid()))

	// 启动接口服务
	webapi.Run(app.ServeApi)

	tool.ExitNotify(func() {
		webapi.Shutdown(app.ServeApi)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
