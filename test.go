package main

import (
	"go-com/config"
	"go-com/core/es"
	"go-com/core/logr"
	"go-com/internal/app"
)

func main() {
	config.Load()
	logr.InitLog("test")

	app.Es8 = es.NewEs8(es.Config{Es: config.C.Es})

	app.Es8.
}
