package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/kafka"
	"go-com/core/logr"
	"time"
)

func main() {
	config.Load()
	logr.InitLog("example")

	k := kafka.Kafka{}
	k.InitProducer(kafka.Config{Kafka: config.C.Kafka})

	k.Produce(nil, []byte("hello 你好1"), "")
	k.Produce(nil, []byte("hello 你好2"), "")
	k.Produce(nil, []byte("hello 你好3"), "")
	k.CloseProducer()

	k.InitConsumer(kafka.Config{Kafka: config.C.Kafka}, "earliest")
	for {
		k.Consume(func(key []byte, msg []byte, timestamp *time.Time) {
			fmt.Println(string(msg))
		})
	}
}
