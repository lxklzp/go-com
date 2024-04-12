package service

import (
	"github.com/sony/gobreaker"
	"go-com/config"
	"time"
)

// 熔断器，错误比例熔断策略
//
// 使用github.com/sony/gobreaker实现的熔断器，因子：
// 状态 state：关闭、半开、打开
// 关闭 ---> readyToTrip ---> 打开
// 打开 ---> timeout ---> 半开
// 半开 ---> onFailure ---> 打开
// 半开 ---> ConsecutiveSuccesses>=maxRequests ---> 关闭
//
// 最大请求次数，MaxRequests：半开时，与Requests、ConsecutiveSuccesses比较
// 一个周期的时长，Interval：关闭时
// 打开到半开持续的时长，Timeout：打开时
// 熔断条件，ReadyToTrip：半开时
//
// 关闭到打开，一个周期内最小请求次数
// 关闭到打开，一个周期内错误百分比，单位%

type RateBreakerConfig struct {
	config.RateBreaker
}

type RateBreaker struct {
	cfg RateBreakerConfig
	cb  *gobreaker.CircuitBreaker
}

func NewRateBreaker(cfg RateBreakerConfig) *RateBreaker {
	st := gobreaker.Settings{
		Name:        "rate_breaker",
		MaxRequests: cfg.HafMaxRequests,
		Interval:    time.Second * time.Duration(cfg.Interval),
		Timeout:     time.Second * time.Duration(cfg.OpenTimeout),
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := counts.TotalFailures * 100 / counts.Requests
			return counts.Requests >= cfg.CloseMinRequests && failureRatio >= cfg.CloseErrorPercent
		},
	}
	rb := &RateBreaker{
		cfg: cfg,
		cb:  gobreaker.NewCircuitBreaker(st),
	}
	return rb
}

// Execute 将要调用的方法包在Execute里面
func (rb *RateBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return rb.cb.Execute(req)
}
