package rds

import (
	"context"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"go-com/config"
	"go-com/core/logr"
	"strconv"
	"time"
)

type Config struct {
	config.Redis
}

const (
	ErrRedisExGetFail = "redis排他锁获取失败。"

	ConnectionTypeSingle   = 1
	ConnectionTypeSentinel = 2
)

func NewRedis(cfg Config) *redis.Client {
	var rds *redis.Client
	// 这是个redis连接池
	switch cfg.Type {
	case ConnectionTypeSingle:
		rds = redis.NewClient(&redis.Options{
			Addr:       cfg.Addr,
			Username:   cfg.Username,
			Password:   cfg.Password,
			DB:         cfg.Db,
			ClientName: cfg.ClientName,
			PoolSize:   cfg.PoolSize,
		})
	case ConnectionTypeSentinel:
		rds = redis.NewFailoverClient(&redis.FailoverOptions{
			Username:   cfg.Username,
			Password:   cfg.Password,
			DB:         cfg.Db,
			ClientName: cfg.ClientName,
			PoolSize:   cfg.PoolSize,

			MasterName:       cfg.MasterName,
			SentinelAddrs:    cfg.SentinelAddrs,
			SentinelUsername: cfg.SentinelUsername,
			SentinelPassword: cfg.SentinelPassword,
		})
	default:
		logr.L.Fatal("不支持的redis连接方式。")
	}

	if _, err := rds.Ping(context.TODO()).Result(); err != nil {
		logr.L.Fatal(err)
	}
	return rds
}

/* ---------- 计数器类型的锁 ---------- */

func lockIndexKey(ty string, seq string) string {
	return config.C.App.Prefix + ":lock-index:" + ty + ":" + seq
}

// IncrLockIndex 增加锁计数器 有效期单位：秒
func IncrLockIndex(r *redis.Client, ty string, seq string, expire int, maxIndexNum int64) bool {
	ctx := context.Background()
	key := lockIndexKey(ty, seq)
	indexNum, err := r.Incr(ctx, key).Result()
	if err != nil {
		logr.L.Error(err.Error())
		return false
	}
	if indexNum == 1 {
		r.Expire(ctx, key, time.Duration(expire)*time.Second)
	}
	if indexNum > maxIndexNum {
		return false
	}
	return true
}

// DelLockIndex 删除锁计数器
func DelLockIndex(r *redis.Client, ty string, seq string) {
	ctx := context.Background()
	key := lockIndexKey(ty, seq)
	r.Del(ctx, key)
}

// GetLockIndex 获取锁计数器
func GetLockIndex(r *redis.Client, ty string, seq string) int {
	ctx := context.Background()
	key := lockIndexKey(ty, seq)
	indexNumStr, _ := r.Get(ctx, key).Result()
	indexNum, _ := strconv.Atoi(indexNumStr)
	return indexNum
}

/* ---------- 排他锁 ---------- */

func lockExKey(ty string) string {
	return config.C.App.Prefix + ":lock-ex:" + ty
}

// LockEx 获取锁
func lockEx(r *redis.Client, ty string, token string, expire int) (bool, error) {
	ctx := context.TODO()
	return r.SetNX(ctx, lockExKey(ty), token, time.Duration(expire)*time.Second).Result()
}

// LockExTry 轮询尝试获取锁
func LockExTry(r *redis.Client, ty string, token string, expire int, period time.Duration) error {
	var ok bool
	var err error
	i := 1
	var maxTry int
	if period != 0 {
		maxTry = int(time.Duration(expire)*time.Second/period) + 2
	}
	for {
		ok, err = lockEx(r, ty, token, expire)
		if err != nil {
			return err
		}
		if !ok {
			if i >= maxTry {
				return errors.New(ErrRedisExGetFail)
			}

			time.Sleep(period)
			i++
		} else {
			return nil
		}
	}
}

// UnLockEx 释放锁
func UnLockEx(r *redis.Client, ty string, token string) (bool, error) {
	ctx := context.TODO()
	// 使用 Lua 脚本确保只有锁的持有者才能释放锁
	script := `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	else
		return 0
	end
	`
	keys := []string{lockExKey(ty)}
	args := []interface{}{token}

	result, err := r.Eval(ctx, script, keys, args).Result()
	if err != nil {
		return false, err
	}

	// Lua 脚本返回 1 表示成功释放锁
	if result == int64(1) {
		return true, nil
	} else {
		return false, nil
	}
}

/* ---------- 常用方法 ---------- */

func Scan(r *redis.Client, keyPrefix string, handler func(key string) error) error {
	ctx := context.TODO()
	var cursor uint64
	var keys []string
	var err error
	for {
		keys, cursor, err = r.Scan(ctx, cursor, keyPrefix+"*", 10).Result()
		if err != nil {
			return err
		}
		for _, key := range keys {
			if err = handler(key); err != nil {
				return err
			}
		}
		if cursor == 0 {
			return nil
		}
	}
}
