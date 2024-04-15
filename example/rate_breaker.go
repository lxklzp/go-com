package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/service"
	"time"
)

func main() {
	config.Load()
	logr.InitLog("test")

	rb := service.NewRateBreaker(service.RateBreakerConfig{RateBreaker: config.C.RateBreaker})
	for i := 0; i < 1000; i++ {
		// 动态修改配置
		if i == 200 {
			rb = service.NewRateBreaker(service.RateBreakerConfig{RateBreaker: config.RateBreaker{
				Interval:          20,
				OpenTimeout:       10,
				HafMaxRequests:    50,
				CloseMinRequests:  5,
				CloseErrorPercent: 30,
			}})
		}
		time.Sleep(time.Millisecond * 100)

		i := i
		go func() {
			result, err := rb.Execute(func() (interface{}, error) {
				if i < 110 && i%2 == 0 {
					return nil, errors.New("错误")
				}

				//if i > 200 && i%2 == 0 {
				//	return nil, errors.New("错误")
				//}
				return 1, nil
			})
			fmt.Println(i, result, err)
		}()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
