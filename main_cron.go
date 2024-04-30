package main

import (
	"context"
	"github.com/robfig/cron/v3"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/system"
	"os"
	"strconv"
)

func main() {
	config.Load()
	logr.InitLog("cront")
	c := cron.New()

	logr.L.Info("启动系统:" + strconv.Itoa(os.Getpid()))
	tool.ExitNotify(func() {
		c.Stop()
	})

	// 定时任务使用示例：统计，每隔6小时执行一次
	_, err := c.AddFunc("1 */6 * * *", func() {
		logr.L.Info("开始统计xxx")
		indexAll := system.Stat()
		logr.L.Infof("完成统计xxx，共计：%d", indexAll)
	})
	if err != nil {
		logr.L.Error(err)
	}

	c.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
