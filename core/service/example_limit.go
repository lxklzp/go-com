package service

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"time"
)

// ExampleLimit 限流算法令牌桶golang标准库time/rate示例：
func ExampleLimit() {
	limit := rate.NewLimiter(3, 5) // 每秒产生 3 个token，桶容量 5

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel() // 超时取消

	for i := 0; ; i++ { // 有多少令牌直接消耗掉
		fmt.Printf("%03d %s\n", i, time.Now().Format("2006-01-02 15:04:05.000"))
		err := limit.Wait(ctx)
		// Wait / WaitN 阻塞，分等级
		// Allow / AllowN 丢弃，分等级
		if err != nil { // 超时取消 err != nil
			fmt.Println("err: ", err.Error())
			return // 超时取消，退出 for
		}
	}
}
