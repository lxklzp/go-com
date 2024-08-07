package main

import (
	"go-com/config"
	"go-com/core/ds"
	"go-com/core/logr"
	"go-com/core/my"
	"go-com/core/pg"
	"go-com/internal/app"
)

func main() {
	config.Load()
	logr.InitLog("create_db_table")
	app.Mysql = my.NewDb(my.Config{Mysql: config.C.Mysql})
	app.Pg = pg.NewDb(pg.Config{Postgresql: config.C.Postgresql})

	pg2my := ds.NewPg2My("postgresql2mysql_multi.yaml", app.Pg, pg.Config{Postgresql: config.C.Postgresql}, app.Mysql, my.Config{Mysql: config.C.Mysql})
	pg2my.Process()
}
