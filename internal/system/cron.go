package system

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/app"
)

func CronRun() {
	l := &cronLog{}
	app.Cron = cron.New(cron.WithLogger(cron.DiscardLogger), cron.WithChain(cron.SkipIfStillRunning(l), cron.Recover(l)))

	var err error
	// 清除过期数据
	_, err = app.Cron.AddFunc("0 1 * * *", func() {
		logr.L.Debug("开始清除过期数据...")
		ClearDaily()
		logr.L.Debug("完成清除过期数据")
	})
	if err != nil {
		logr.L.Error(err)
	}

	app.Cron.Start()
}

type cronLog struct {
}

func (l *cronLog) Info(msg string, keysAndValues ...interface{}) {
	logr.L.WithFields(logrus.Fields{
		"data": keysAndValues,
	}).Info(msg)
}

func (l *cronLog) Error(err error, msg string, keysAndValues ...interface{}) {
	tool.ErrorStack(err)
}

func ClearDaily() {
}
