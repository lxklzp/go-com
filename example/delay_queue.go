package main

import (
	"context"
	"fmt"
	"go-com/config"
	"go-com/core/dq"
	"go-com/core/logr"
)

func main() {
	config.Load()
	logr.InitLog("example")

	q := dq.NewQueue()
	go q.Run(func(m dq.Message) {
		fmt.Println(m)
	})

	//q.Produce(dq.Message{
	//	Timestamp: time.Now().Unix(),
	//	Topic:     "test",
	//	No:        1,
	//	Data:      "1111111",
	//})
	//q.Produce(dq.Message{
	//	Timestamp: time.Now().Unix() - 100,
	//	Topic:     "test2",
	//	No:        2,
	//	Data:      "22222",
	//})
	//q.Produce(dq.Message{
	//	Timestamp: time.Now().Unix() + 30,
	//	Topic:     "test",
	//	No:        3,
	//	Data:      "3333",
	//})
	//q.Produce(dq.Message{
	//	Timestamp: time.Now().Unix() + 12,
	//	Topic:     "test2",
	//	No:        4,
	//	Data:      "4444",
	//})
	//q.Produce(dq.Message{
	//	Timestamp: time.Now().Unix() - 6,
	//	Topic:     "test",
	//	No:        5,
	//	Data:      "55555",
	//})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
