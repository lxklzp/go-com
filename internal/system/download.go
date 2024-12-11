package system

import (
	"github.com/pkg/errors"
	"go-com/config"
	"go-com/core/logr"
	"go-com/internal/app"
	"go-com/internal/model"
	"sync"
	"time"
)

// 最大并行下载数量
const downloadLimitMax = 3

var downloadLimit int
var downloadLimitLock sync.Mutex

func DownloadBefore(title string, userId int) (int, error) {
	downloadLimitLock.Lock()
	if downloadLimit >= downloadLimitMax {
		downloadLimitLock.Unlock()
		return 0, errors.New("当前有其它列表正在下载，请稍后重试。")
	}
	downloadLimit++
	downloadLimitLock.Unlock()

	now := config.Timestamp(time.Now())
	m := model.Download{
		Name:       title,
		Path:       "",
		UserID:     userId,
		BeginTime:  now,
		EndTime:    config.DefaultTimeMin,
		CreateTime: now,
		Status:     1,
	}
	app.Pg.Create(&m)
	return m.ID, nil
}

func DownloadAfter(id int, path string, err error) {
	downloadLimitLock.Lock()
	downloadLimit--
	downloadLimitLock.Unlock()
	now := config.Timestamp(time.Now())
	if err != nil {
		logr.L.Error(err)
		app.Pg.Where("id=?", id).Updates(model.Download{Status: 3, EndTime: now})
		return
	}
	app.Pg.Where("id=?", id).Updates(model.Download{Status: 2, EndTime: now, Path: path})
}
