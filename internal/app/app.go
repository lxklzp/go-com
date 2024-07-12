package app

import (
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-redis/redis/v8"
	"github.com/robfig/cron/v3"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"go-com/core/kafka"
	"go-com/core/service"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

var Clickhouse *gorm.DB
var Pg *gorm.DB
var Mysql *gorm.DB
var Redis *redis.Client
var Etcd *clientv3.Client
var Nb *nebula.SessionPool
var SD *service.Discovery
var Es *elasticsearch.Client
var KafkaP kafka.Kafka
var KafkaCQ kafka.Kafka
var Cron *cron.Cron
