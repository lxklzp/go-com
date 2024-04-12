package service

import (
	"context"
	"github.com/pkg/errors"
	"go-com/config"
	"golang.org/x/time/rate"
	"time"
)

// 两层限流：外层令牌桶，内层chan
//
// 令牌桶：
// 使用golang.org/x/time/rate实现的令牌桶，因子：
// 桶中每秒产生token个数（产生一个token的时长），Limit
// 桶中token最大个数，Burst
// 桶中token耗尽后等待刷新token时长，Timeout
//
// chan：
// 桶外token存量最大个数，Stock

type RateLimitConfig struct {
	config.RateLimit
}

type RateLimit struct {
	cfg     RateLimitConfig
	limiter *rate.Limiter
	stockCh chan bool
}

func NewRateLimit(cfg RateLimitConfig) *RateLimit {
	rl := &RateLimit{
		cfg:     cfg,
		limiter: rate.NewLimiter(rate.Limit(cfg.Limit), cfg.Burst),
		stockCh: make(chan bool, cfg.Stock),
	}
	return rl
}

func (rl *RateLimit) LimitBefore() error {
	var err error
	if rl.cfg.Timeout != 0 {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*time.Duration(rl.cfg.Timeout))
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
	if rl.cfg.Stock == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*time.Duration(rl.cfg.Timeout))
	defer cancel()
	select {
	case <-ctx.Done():
		return errors.New("存量数目超过限制，已限流")
	case rl.stockCh <- true:
		return nil
	}
}

func (rl *RateLimit) LimitAfter() {
	if rl.cfg.Stock == 0 {
		return
	}
	<-rl.stockCh
}
