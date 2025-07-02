package system

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/mod"
	"go-com/core/rds"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/model"
	"sync/atomic"
	"time"
)

func CronRun() {
	l := &cronLog{}
	app.Cron = cron.New(cron.WithLogger(cron.DiscardLogger), cron.WithChain(cron.Recover(l)))

	var err error
	// 清除过期的数据
	_, err = app.Cron.AddFunc("0 1 * * *", func() {
		logr.L.Debug("开始清除过期的数据...")
		if err := rds.LockExTry(app.Redis, "CronClearHistory", "CronClearHistory", 10, 0); err != nil {
			if err.Error() != rds.ErrRedisExGetFail {
				logr.L.Error(err.Error())
			}
			return
		}
		defer func() {
			if _, err := rds.UnLockEx(app.Redis, "CronClearHistory", "CronClearHistory"); err != nil {
				logr.L.Error(err.Error())
			}
		}()

		CronClearHistory()
		logr.L.Debug("完成清除过期数据")
	})
	if err != nil {
		logr.L.Error(err)
	}

	app.Cron.Start()
	logr.L.Info("定时任务启动成功")
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

// 按表清除过期的数据
func clearTable(table string, primaryKey string, createTimeName string, limitTime string, pageSize int) {
	var m mod.PrimaryId
	app.Db.Table(table).Select(primaryKey).Where(createTimeName+"<?", limitTime).Take(&m)
	for m.ID > 0 {
		app.Db.Table(table).Where(createTimeName+"<?", limitTime).Limit(pageSize).Delete(nil)
		m = mod.PrimaryId{}
		app.Db.Table(table).Select(primaryKey).Where(createTimeName+"<?", limitTime).Take(&m)
	}
}

var CronClearHistoryLock atomic.Int64

// CronClearHistory 清除过期的数据
func CronClearHistory() {
	// 同时只能有一个方法运行
	if !CronClearHistoryLock.CompareAndSwap(0, 1) {
		return
	}
	defer CronClearHistoryLock.Store(0)

	limitT := time.Now().AddDate(0, 0, -7)
	limitTime := limitT.Format(config.DateTimeFormatter)
	pageSize := 10000

	clearTable((&model.Download{}).TableName(), "id", "create_time", limitTime, pageSize)
}
