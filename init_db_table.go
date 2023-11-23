package main

import (
	"go-com/config"
	"go-com/global"
	"go-com/internal/model"
)

func main() {
	config.Load()
	global.InitDefine()
	global.InitLog("init_db_table")
	global.InitGormPg()

	if err := global.GormPg.Exec("drop table if exists rule_storage").Error; err != nil {
		global.Log.Fatal(err)
	}
	if err := global.GormPg.AutoMigrate(&model.RuleStorage{}); err != nil {
		global.Log.Fatal(err)
	}
}
