package main

import (
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/security"
)

func main() {
	config.Load()
	logr.InitLog("create_db_table")

	security.GenerateCert([]string{"127.0.0.1"})
}
