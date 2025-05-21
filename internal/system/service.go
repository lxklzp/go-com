package system

import (
	"go-com/config"
	"go-com/core/kafka"
	"go-com/core/my"
	"go-com/core/orm"
	"go-com/core/pg"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/system/resource"
	"time"
)

func RunApp() {
	run()

	app.KafkaP.InitProducer(kafka.Config{Kafka: config.C.KafkaP})
	kafkaC()

	CronRun()
}

func RunWeb() {
	config.C.App.Id += 100 // app与web共用同一个id配置，这里web的id自增100以作区分
	run()

	resource.Download.Init()
}

func run() {
	switch config.C.App.DbType {
	case orm.DbMysql:
		app.Db = my.NewDb(my.Config{Mysql: config.C.Mysql})
	case orm.DbPgsql:
		app.Db = pg.NewDb(pg.Config{Postgresql: config.C.Pg})
	}

	tool.InitSnowflake()
}

func kafkaC() {
	cfg := kafka.Config{Kafka: config.C.KafkaC}
	for i := 0; i < cfg.MaxConsumeWorkerNum; i++ {
		go func() {
			var client kafka.Kafka
			client.InitConsumer(cfg, "latest")
			for {
				client.Consume(func(key []byte, msg []byte, timestamp *time.Time) {
					// 业务逻辑
				})
			}
		}()
	}
}
