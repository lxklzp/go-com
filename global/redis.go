package global

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go-com/config"
)

var Redis *redis.Client

func InitRedis() {
	cfg := config.C.Redis
	// 这是个redis连接池
	Redis = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.Db,
		PoolSize: 10,
	})
	if _, err := Redis.Ping(context.TODO()).Result(); err != nil {
		Log.Fatal(err)
	}
}
