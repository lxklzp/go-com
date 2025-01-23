package system

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/app"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
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

func ClearDaily() {
	var err error
	now := time.Now()

	/* 导出的excel文件 */
	var dir string
	prevDay := 5
	// 删除上月数据
	if now.Day() == prevDay {
		dir = config.C.App.PublicPath + "/export/" + now.AddDate(0, -1, 0).Format(config.MonthNumberFormatter)
		os.RemoveAll(dir)
	}
	// 删除5天前数据
	dir = config.C.App.PublicPath + "/export/" + now.Format(config.MonthNumberFormatter)
	dateNumber := now.AddDate(0, 0, -prevDay).Format(config.DateNumberFormatter)
	err = filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 跳过目录自身
		if path == dir {
			return nil
		}
		filename := info.Name()
		if info.IsDir() {
			return nil
		} else if strings.Contains(filename, dateNumber) {
			return os.Remove(dir + "/" + filename)
		}
		return err
	})
	if err != nil {
		logr.L.Error(err)
	}
}
