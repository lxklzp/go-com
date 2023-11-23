package main

import (
	"fmt"
	"go-com/config"
	"go-com/global"
	"os"
	"time"
)

func main() {
	config.Load()
	global.InitLog("kafka_com")

	var consumer global.Kafka
	cfg := config.C.Kafka
	consumer.InitConsumer(cfg, "earliest")

	for {
		consumer.Consume(func(msg []byte, timestamp *time.Time) {
			fmt.Printf(string(msg))
			os.Exit(0)
		})
	}
}
