package main

import (
	"context"
	"fmt"
	"go-com/config"
	"go-com/global"
)

func main() {
	config.Load()
	global.InitDefine()
	global.InitLog("hello")

	l, _ := global.IPString2Long("192.168.132.1")
	fmt.Println(global.Long2IPString(l))

	// 保持主协程
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
