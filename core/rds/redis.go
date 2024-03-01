package rds

import (
	"context"
	"github.com/go-redis/redis/v8"
	"go-com/core/logr"
)

type Config struct {
	Addr     string
	Password string
	Db       int
}

func NewRedis(cfg Config) *redis.Client {
	// 这是个redis连接池
	rds := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.Db,
		PoolSize: 10,
	})
	if _, err := rds.Ping(context.TODO()).Result(); err != nil {
		logr.L.Fatal(err)
	}
	return rds
}
