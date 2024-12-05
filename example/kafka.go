package main

import (
	"go-com/config"
	"go-com/core/kafka"
	"go-com/internal/app"
	"time"
)

func InitSystem() {
	app.KafkaP.InitProducer(kafka.Config{Kafka: config.C.Kafka})
	go exampleResp()
}

var KafkaTopicExampleResp = "example_resp"

// 消费kafka
func exampleResp() {
	cfg := kafka.Config{Kafka: config.C.Kafka}

	for i := 0; i < cfg.MaxConsumeWorkerNum; i++ {
		go func() {
			var client kafka.Kafka
			client.InitConsumer(cfg, "latest")
			for {
				client.Consume(func(key []byte, msg []byte, timestamp *time.Time) {
					// TODO
				})
			}
		}()
	}
}
