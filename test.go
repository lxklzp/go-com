package main

import (
	"context"
	"fmt"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/pg"
	"go-com/core/tool"
	"go-com/internal/app"
)

func main() {
	config.Load()
	logr.InitLog("web")
	app.Pg = pg.NewDb(pg.Config{Postgresql: config.C.Postgresql})

	fmt.Println(tool.InArray([]int{1, 2, 3}, 3))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
