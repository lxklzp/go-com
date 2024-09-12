package system

import (
	"go-com/core/logr"
	"go-com/internal/app"
)

func CronRun() {
	var err error
	// 清除过期数据
	_, err = app.Cron.AddFunc("0 1 * * *", func() {
		logr.L.Info("开始清除过期数据")
		ClearDaily()
		logr.L.Infof("完成清除过期数据")
	})
	if err != nil {
		logr.L.Error(err)
	}

	app.Cron.Start()
}

func ClearDaily() {
}
