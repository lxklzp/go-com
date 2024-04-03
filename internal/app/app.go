package app

import (
	"github.com/go-redis/redis/v8"
	"go-com/core/service"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gorm.io/gorm"
	"net/http"
	"net/http/httputil"
)

var ServeApi *http.Server
var Clickhouse *gorm.DB
var Pg *gorm.DB
var Mysql *gorm.DB
var Redis *redis.Client
var Etcd *clientv3.Client
var SD *service.Discovery
var ProxyVmc *httputil.ReverseProxy
