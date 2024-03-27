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
	logr.InitLog("test")

	k := kafka.Kafka{}
	//k.InitProducer(kafka.Config{Kafka: config.C.Kafka})
	//
	//k.Produce([]byte("hello 你好2"))
	//k.Produce([]byte("hello 你好3"))
	//k.Produce([]byte("hello 你好4"))
	//k.CloseProducer()

	k.InitConsumer(kafka.Config{Kafka: config.C.Kafka}, "earliest")
	for {
		k.Consume(func(msg []byte, timestamp *time.Time) {
			fmt.Println(string(msg))
		})
	}
}
