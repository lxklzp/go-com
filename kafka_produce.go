package main

import (
	"go-com/config"
	"go-com/global"
)

func main() {
	config.Load()
	global.InitLog("kafka_produce")
	var produce global.Kafka
	cfg := config.C.Kafka
	cfg.Topic = "new-sealhead-zftalarm"
	produce.Produce(cfg, []byte(``))
}
