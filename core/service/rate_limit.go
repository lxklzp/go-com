package service

import (
	"context"
	"github.com/pkg/errors"
	"go-com/config"
	"golang.org/x/time/rate"
	"sync/atomic"
	"time"
)

// 两层限流：外层令牌桶，内层chan，支持动态更新配置参数
//
// 令牌桶：
// 使用golang.org/x/time/rate实现的令牌桶，因子：
// 桶中每秒产生token个数（产生一个token的时长），Limit
// 桶中token最大个数，Burst
// 桶中token耗尽后等待刷新token时长，Timeout
//
// chan：
// 桶外token存量最大个数，MaxStock

type RateLimitConfig struct {
	config.RateLimit
}

type RateLimit struct {
	limiter  *rate.Limiter
	maxStock atomic.Int32
	stock    atomic.Int32
	timeout  atomic.Int32
}

func NewRateLimit(cfg RateLimitConfig) *RateLimit {
	rl := &RateLimit{
		limiter: rate.NewLimiter(rate.Limit(cfg.Limit), cfg.Burst),
	}
	rl.maxStock.Store(cfg.MaxStock)
	rl.stock.Store(0)
	rl.timeout.Store(cfg.Timeout)
	return rl
}

func (rl *RateLimit) Entry() error {
	var err error
	timeout := rl.timeout.Load()
	if timeout != 0 {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*time.Duration(timeout))
		defer cancel()
		if err = rl.limiter.Wait(ctx); err != nil {
			return err
		} else {
			return rl.limitStock()
		}
	} else {
		if rl.limiter.Allow() {
			return rl.limitStock()
		} else {
			return errors.New("并发数目超过限制，已限流")
		}
	}
}

func (rl *RateLimit) limitStock() error {
	stock := rl.stock.Add(int32(1))
	maxStock := rl.maxStock.Load()
	if maxStock == 0 {
		return nil
	}
	if stock > maxStock {
		rl.stock.Add(int32(-1))
		return errors.New("存量请求数目超过限制，已限流")
	}
	return nil
}

func (rl *RateLimit) Exit() {
	rl.stock.Add(int32(-1))
}

func (rl *RateLimit) SetConfig(cfg RateLimitConfig) {
	rl.limiter.SetLimit(rate.Limit(cfg.Limit))
	rl.limiter.SetBurst(cfg.Burst)
	rl.timeout.Store(cfg.Timeout)
	rl.maxStock.Store(cfg.MaxStock)
}
