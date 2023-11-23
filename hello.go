package main

import (
	"context"
	"go-com/config"
	"go-com/global"
)

func main() {
	config.Load()
	global.InitDefine()
	global.InitLog("hello")
	global.InitGormPg()

	// 保持主协程
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
