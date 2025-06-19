package app

import (
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/redis/go-redis/v9"
	"github.com/robfig/cron/v3"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"go-com/core/kafka"
	"go-com/core/service"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
)

var Cron *cron.Cron

var Db *gorm.DB
var Redis *redis.Client
var KafkaP kafka.Kafka
var Es *elasticsearch.Client
var Etcd *clientv3.Client
var Clickhouse *gorm.DB
var Nb *nebula.SessionPool
var SD *service.Discovery
