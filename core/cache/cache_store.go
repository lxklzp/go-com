package cache

import (
	"go-com/core/logr"
	"go-com/internal/app"
	"sync"
)

// 缓存存储 利用缓存，将数据定时或定量的批量存储

var CacheStore cacheStore

type cacheStore struct {
	lock   sync.Mutex
	list   []interface{}
	status int // 1 空闲 2 保存中
}

const (
	CacheStoreCacheMax = 5000

	CacheStoreStatusFree  = 0
	CacheStoreStatusStore = 1
)

// Append 追加
func (al *cacheStore) Append(item interface{}) {
	al.lock.Lock()
	defer al.lock.Unlock()

	if len(al.list) >= CacheStoreCacheMax {
		// 缓存已满，并且状态为存储中，丢弃
		if al.status == CacheStoreStatusStore {
			return
		}

		al.storeAndReset()
	}
	al.list = append(al.list, item)
}

// 存储并重置缓存
func (al *cacheStore) storeAndReset() {
	if len(al.list) == 0 {
		return
	}

	list := al.list
	al.list = make([]interface{}, 0, CacheStoreCacheMax)
	al.status = CacheStoreStatusStore

	go func() {
		al.store(list, false)
	}()
}

// store 存储
func (al *cacheStore) store(list []interface{}, isExit bool) {
	var err error
	if err = app.Mysql.Create(list).Error; err != nil {
		logr.L.Error(err)
	}
	if isExit {
		return
	}

	al.lock.Lock()
	defer al.lock.Unlock()
	al.status = CacheStoreStatusFree
}

// StoreForCron 供定时任务调用
func (al *cacheStore) StoreForCron() {
	al.lock.Lock()
	defer al.lock.Unlock()

	al.storeAndReset()
}

// StoreForExit 程序退出时调用
func (al *cacheStore) StoreForExit() {
	al.lock.Lock()
	defer al.lock.Unlock()

	if len(al.list) == 0 {
		return
	}

	al.store(al.list, true)
}
