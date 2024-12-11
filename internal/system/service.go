package system

import (
	"go-com/config"
	"go-com/core/kafka"
	"go-com/internal/app"
	"time"
)

func Run() {
	ShiLian.Init()

	app.KafkaP.InitProducer(kafka.Config{Kafka: config.C.KafkaP})
	kafkaC()
	CronRun()
}

func kafkaC() {
	cfg := kafka.Config{Kafka: config.C.KafkaC}

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
