package main

import (
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/pg"
	"go-com/internal/app"
)

func main() {
	config.Load()
	logr.InitLog("create_db_table")

	app.Pg = pg.NewDb(pg.Config{Postgresql: config.C.Postgresql})
	//app.Pg.Table("hehe").Create(map[string]interface{}{
	//	"name":        "bb",
	//	"age":         "2",
	//	"id":          "11",
	//	"user_id":     "1",
	//	"create_time": "2024-08-19 10:10:00",
	//	"money":       "29.12",
	//	"ab":          "111.222",
	//})
	app.Pg.Exec("copy hehe from 'C:/Users/dawnmn/Desktop/hehe.csv' WITH CSV DELIMITER '|'")

}
