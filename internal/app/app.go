package app

import (
	"gorm.io/gorm"
	"net/http"
)

var ServeApi *http.Server
var Pg *gorm.DB
