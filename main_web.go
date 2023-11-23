package main

import (
	"context"
	"go-com/config"
	"go-com/global"
	webApi "go-com/internal/webapi"
	"go-com/lib/service"
	"os"
	"strconv"
)

func main() {
	config.Load()
	global.InitLog("main_web")
	global.InitDefine()
	global.InitGormPg()
	global.InitSnowflake()

	global.Log.Debug("启动系统:" + strconv.Itoa(os.Getpid()))

	if config.C.App.IsDistributed {
		global.InitEtcd()
		go service.SD.Watch(service.SDApiPrefix)
	}

	webApi.Run()

	global.ExitNotify(func() {
		webApi.Shutdown()
		if config.C.App.IsDistributed {
			global.Etcd.Close()
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
