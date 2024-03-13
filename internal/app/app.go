package app

import (
	"gorm.io/gorm"
	"net/http"
	"net/http/httputil"
)

var ServeApi *http.Server
var Pg *gorm.DB
var ProxyVmc *httputil.ReverseProxy
