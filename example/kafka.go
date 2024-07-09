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
	cfg.Topic = KafkaTopicExampleResp
	cfg.Group = KafkaTopicExampleResp + "_group"
	app.KafkaCQ.InitConsumer(cfg, "latest")
	for {
		app.KafkaCQ.Consume(func(key []byte, msg []byte, timestamp *time.Time) {
			// TODO
		})
	}
}
