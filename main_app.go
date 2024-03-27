package main

import "C"
import (
	"context"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/pg"
	"go-com/core/tool"
	"go-com/internal/api"
	"go-com/internal/app"
	"os"
	"strconv"
)

func main() {
	config.Load()
	logr.InitLog("app")
	app.Pg = pg.NewDb(pg.Config{Postgresql: config.C.Postgresql})
	logr.L.Info("启动系统:" + strconv.Itoa(os.Getpid()))

	// 启动接口服务
	api.Run(app.ServeApi)

	tool.ExitNotify(func() {
		api.Shutdown(app.ServeApi)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
