package resource

import (
	"github.com/pkg/errors"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/model"
	"sync/atomic"
	"time"
)

var Download downloadHandler

const (
	downloadLimitMax = 3 // 最大并行下载数量

	DownloadStatusRunning = 1
	DownloadStatusSuccess = 2
	DownloadStatusFail    = 3
)

type downloadHandler struct {
	downloadLimitLock *atomic.Int64
}

func (handler *downloadHandler) Init() {
	handler.downloadLimitLock = &atomic.Int64{}
}

func (handler *downloadHandler) Before(title string, userId int) (int, error) {
	if !tool.AtomicIncr(handler.downloadLimitLock, 1, downloadLimitMax) {
		return 0, errors.New("下载任务已满，请稍后重试。")
	}

	now := tool.Timestamp(time.Now())
	m := model.Download{
		Name:       title,
		Path:       "",
		UserID:     userId,
		BeginTime:  now,
		CreateTime: now,
		Status:     DownloadStatusRunning,
	}
	app.Db.Omit("end_time").Create(&m)
	return m.ID, nil
}

func (handler *downloadHandler) After(id int, path string, err error) {
	tool.AtomicDecr(handler.downloadLimitLock, 1, 0)
	now := tool.Timestamp(time.Now())
	if err != nil {
		logr.L.Error(err)
		app.Db.Where("id=?", id).Updates(model.Download{Status: DownloadStatusFail, EndTime: now})
		return
	}
	app.Db.Where("id=?", id).Updates(model.Download{Status: DownloadStatusSuccess, EndTime: now, Path: path})
}
