package main

import (
	"context"
	"fmt"
	"go-com/config"
	"go-com/core/etcd"
	"go-com/core/logr"
	"go-com/core/pg"
	"go-com/core/service"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/webapi"
	"os"
	"strconv"
	"strings"
)

func main() {
	config.Load()
	logr.InitLog("web")
	app.Pg = pg.NewDb(pg.Config{Postgresql: config.C.Postgresql})
	logr.L.Info("启动系统:" + strconv.Itoa(os.Getpid()))
	tool.ExitNotify(func() {
		webapi.Shutdown()
	})

	if config.C.App.IsDistributed {
		app.Etcd = etcd.NewEtcd(etcd.Config{Etcd: config.C.Etcd})
		app.SD = service.NewServiceDiscovery(app.Etcd)
		url := fmt.Sprintf("http://%s:%s", config.C.App.PublicIp, strings.Split(config.C.App.WebApiAddr, ":")[1])
		go app.SD.Registry("app", strconv.Itoa(int(config.C.App.Id)), url)
	}

	// 启动接口服务
	webapi.Run()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
