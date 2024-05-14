package main

import (
	"context"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/service"
	"time"
)

func main() {
	config.Load()
	logr.InitLog("example")

	rl := service.NewRateLimit(service.RateLimitConfig{RateLimit: config.C.RateLimit})

	for i := 0; i < 200; i++ {
		if i == 50 {
			// 动态修改配置
			rl.SetConfig(service.RateLimitConfig{RateLimit: config.RateLimit{
				Limit:    10,
				Burst:    20,
				Timeout:  200,
				MaxStock: 5,
			}})
		} else if i == 100 {
			rl.SetConfig(service.RateLimitConfig{RateLimit: config.RateLimit{
				Limit:    4,
				Burst:    5,
				Timeout:  1000,
				MaxStock: 10,
			}})
		} else if i == 150 {
			rl.SetConfig(service.RateLimitConfig{RateLimit: config.RateLimit{
				Limit:    4,
				Burst:    5,
				Timeout:  1000,
				MaxStock: 0,
			}})
		}
		i := i
		time.Sleep(time.Millisecond * 100)
		go func() {
			err := rl.Entry()
			if err != nil {
				logr.L.Error(err)
			} else {
				defer rl.Exit()
				time.Sleep(time.Second * 5)
				logr.L.Info(" -----", i, " -----", time.Now().String())
			}
		}()

	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
