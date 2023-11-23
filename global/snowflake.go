package global

import (
	"go-com/config"
	"sync"
	"time"
)

const (
	snowflakeWorkerBits  uint8 = 10
	snowflakeNumberBits  uint8 = 12
	snowflakeWorkerMax   int64 = -1 ^ (-1 << snowflakeWorkerBits)
	snowflakeNumberMax   int64 = -1 ^ (-1 << snowflakeNumberBits)
	snowflakeTimeShift   uint8 = snowflakeWorkerBits + snowflakeNumberBits
	snowflakeWorkerShift uint8 = snowflakeNumberBits
	snowflakeEpoch       int64 = 1682524800000 // 2023-04-27 00:00:00 上线后不能修改，否则会生成相同的ID
)

var SnowflakeMergeGroup snowflakeWorker
var SnowflakeFbOrder snowflakeWorker

type snowflakeWorker struct {
	mu        sync.Mutex
	timestamp int64
	workerId  int64
	number    int64
}

func InitSnowflake() {
	if config.C.App.Id < 0 || config.C.App.Id > snowflakeWorkerMax {
		Log.Fatalf("服务程序唯一编号Id有误：%d", config.C.App.Id)
	}
	SnowflakeMergeGroup = snowflakeWorker{
		timestamp: 0,
		workerId:  config.C.App.Id,
		number:    0,
	}
	SnowflakeFbOrder = snowflakeWorker{
		timestamp: 0,
		workerId:  config.C.App.Id,
		number:    0,
	}
}

func (w *snowflakeWorker) GetId() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now().UnixMilli()
	if w.timestamp == now {
		w.number++
		if w.number > snowflakeNumberMax {
			for now <= w.timestamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		w.number = 0
		w.timestamp = now
	}
	ID := (now-snowflakeEpoch)<<snowflakeTimeShift | (w.workerId << snowflakeWorkerShift) | (w.number)
	return ID
}
