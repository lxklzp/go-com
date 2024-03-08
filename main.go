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
	logr.InitLog("main")
	app.Pg = pg.NewDb(pg.Config{Postgresql: config.C.Postgresql})

	logr.L.Debug("启动系统:" + strconv.Itoa(os.Getpid()))

	api.Run(app.ServeApi)

	tool.ExitNotify(func() {
		api.Shutdown(app.ServeApi)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
