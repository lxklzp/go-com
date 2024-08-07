package main

import (
	"context"
	"github.com/robfig/cron/v3"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/system"
	"os"
	"strconv"
)

func main() {
	config.Load()
	logr.InitLog("cront")
	app.Cron = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DefaultLogger)))

	logr.L.Info("启动系统:" + strconv.Itoa(os.Getpid()))
	tool.ExitNotify(func() {
		app.Cron.Stop()
	})

	cronRun()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}

func cronRun() {
	// 定时任务使用示例
	_, err := app.Cron.AddFunc("* * * * *", func() {
		logr.L.Info("开始统计xxx")
		indexAll := system.Stat()
		logr.L.Infof("完成统计xxx，共计：%d", indexAll)
	})
	if err != nil {
		logr.L.Error(err)
	}

	app.Cron.Start()
}
