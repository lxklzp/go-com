package main

import (
	"context"
	"go-com/config"
	"go-com/global"
	"go-com/internal/api"
	"os"
	"strconv"
)

func main() {
	config.Load()
	global.InitLog("main")
	global.InitDefine()
	//global.InitGormPg()
	global.InitSnowflake()
	global.Log.Debug("启动系统:" + strconv.Itoa(os.Getpid()))
	//global.InitEtcd()

	//if config.C.App.IsDistributed {
	//	go service.SD.Registry(service.SDApiPrefix, strconv.FormatInt(config.C.App.Id, 10), config.C.App.AppApiAddr)
	//	go service.SD.Registry(service.SDMergePrefix, strconv.FormatInt(config.C.App.Id, 10), config.C.App.AppApiAddr)
	//	go service.SD.Watch(service.SDMergePrefix)
	//}

	api.Run()

	global.ExitNotify(func() {
		api.Shutdown()
		if config.C.App.IsDistributed {
			global.Etcd.Close()
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
