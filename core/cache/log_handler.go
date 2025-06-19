package cache

import (
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"go-com/config"
	"go-com/core/es"
	"go-com/core/logr"
	"go-com/core/tool"
	"sync"
	"sync/atomic"
	"time"
)

// 日志处理器，缓存数据，定时（同步）定量（异步）保存

var Log logHandler

// 日志单元
type log struct {
	ty   string     // 日志分类
	lock sync.Mutex // 缓存锁

	logList    []map[string]interface{} // 缓存的日志列表
	saveStatus int                      // 日志保存状态：1 空闲 2 保存中
}

// 日志处理器
type logHandler struct {
	metaLogList map[string]*log // 日志单元列表
	savingCount *atomic.Int64   // 异步保存时，正在执行保存到es的协程数，超过的丢弃

	saveLock atomic.Int64 // 定时任务锁

	esClient *elasticsearch.Client
	esPrefix string
	prefix   string
}

const (
	LogCacheMax = 5000 // 定量（异步）保存

	LogStatusFree  = 1
	LogStatusStore = 2

	MaxSavingCount = 10 // 异步保存时，正在执行保存到es的最大协程数
)

// GetIndex 获取日志数据的索引
func (handler *logHandler) getEsIndex(ty string) string {
	return fmt.Sprintf("%s%s_%s", handler.esPrefix, handler.prefix, ty)
}

func (handler *logHandler) init(esClient *elasticsearch.Client, esPrefix string, prefix string, typeList []string) {
	handler.esClient = esClient
	handler.esPrefix = esPrefix
	handler.prefix = prefix
	handler.metaLogList = make(map[string]*log)
	handler.savingCount = &atomic.Int64{}

	// 初始化日志单元列表
	for _, ty := range typeList {
		handler.metaLogList[ty] = &log{
			ty:         ty,
			logList:    make([]map[string]interface{}, 0, LogCacheMax),
			saveStatus: LogStatusFree,
		}
	}
}

// 设置日志单元状态
func (handler *logHandler) setMetaLogStatus(key string, saveStatus int) {
	handler.metaLogList[key].lock.Lock()
	defer handler.metaLogList[key].lock.Unlock()

	handler.metaLogList[key].saveStatus = saveStatus
}

// 追加日志
func (handler *logHandler) appendLog(ty string, log map[string]interface{}) {
	if len(log) == 0 {
		return
	}

	key := handler.getEsIndex(ty)
	handler.metaLogList[key].lock.Lock()
	defer handler.metaLogList[key].lock.Unlock()

	if len(handler.metaLogList[key].logList) >= LogCacheMax {
		// 缓存已满，并且状态为存储中，丢弃
		if handler.metaLogList[key].saveStatus == LogStatusStore {
			return
		}

		// 缓存已满，并且状态为空闲

		// 重置缓存
		logList := handler.metaLogList[key].logList
		handler.metaLogList[key].logList = make([]map[string]interface{}, 0, LogCacheMax)
		handler.metaLogList[key].saveStatus = LogStatusStore

		// 异步保存到es
		go func() {
			defer func() {
				handler.setMetaLogStatus(key, LogStatusFree)
			}()

			// 超过正在执行保存到es的最大协程数，丢弃本轮数据
			if !tool.AtomicIncr(handler.savingCount, 1, MaxSavingCount) {
				return
			}
			defer tool.AtomicDecr(handler.savingCount, 1, 0)

			// 写入@timestamp值
			now := time.Now().Format(config.DateTimeStandardZoneFormatter)
			for k := range logList {
				logList[k]["@timestamp"] = now
			}

			logr.L.Debug("定量（异步）保存：", key)
			es.BatchInsert(handler.esClient, key, logList)
		}()
	}

	// 追加日志到缓存
	handler.metaLogList[key].logList = append(handler.metaLogList[key].logList, log)
}

// 定时（同步）保存所有缓存的日志到es，供定时任务调用
func (handler *logHandler) saveLogAll() {
	// 同时只能有一个方法运行
	if !handler.saveLock.CompareAndSwap(0, 1) {
		return
	}
	defer handler.saveLock.Store(0)

	for key := range handler.metaLogList {

		handler.saveLog(key)
	}
}

// 保存缓存的日志到es
func (handler *logHandler) saveLog(key string) {
	handler.metaLogList[key].lock.Lock()

	if len(handler.metaLogList[key].logList) == 0 {
		handler.metaLogList[key].lock.Unlock()
		return
	}

	logList := handler.metaLogList[key].logList

	// 写入@timestamp值
	now := time.Now().Format(config.DateTimeStandardZoneFormatter)
	for k := range logList {
		logList[k]["@timestamp"] = now
	}

	// 重置缓存
	handler.metaLogList[key].logList = make([]map[string]interface{}, 0, LogCacheMax)
	handler.metaLogList[key].saveStatus = LogStatusStore

	handler.metaLogList[key].lock.Unlock()

	defer func() {
		handler.setMetaLogStatus(key, LogStatusFree)
	}()

	// 同步保存到es
	logr.L.Debug("定时（同步）保存：", key)
	es.BatchInsert(handler.esClient, key, logList)
}
