package app

import (
	"gorm.io/gorm"
	"net/http"
	"net/http/httputil"
)

var ServeApi *http.Server
var Clickhouse *gorm.DB
var Pg *gorm.DB
var Mysql *gorm.DB
var ProxyVmc *httputil.ReverseProxy
