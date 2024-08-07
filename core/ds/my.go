package ds

import (
	"fmt"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/my"
	"go-com/core/tool"
	"os"
	"strings"
)

/*---------- mysql 源 ----------*/

// 读取表数据前的准备工作
func myBeforeReadSingleTable(src *Src, tableName string) {
	var err error
	var sql string
	var rows []map[string]interface{}
	if src.Where == "" {
		// 查询表数据总数，预估值
		sql = "SELECT TABLE_ROWS FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA='" + src.Cfg["dbname"] + "' AND TABLE_NAME='" + tableName + "'"
		src.Db.Raw(sql).Scan(&rows)
		if len(rows) == 0 {
			logr.L.Panicf("源库中%s表不存在", tableName)
		}
		src.Count = rows[0]["TABLE_ROWS"].(int64)
	} else {
		// 查询同步数据总数
		if err = src.Db.Table(tableName).Where(src.Where).Count(&src.Count).Error; err != nil {
			logr.L.Panic(err)
		}
	}
	// 构建字段sql
	src.FieldSql = "*"
	if len(src.Field) == 0 {
		// 如果计划没有配置字段，从表中获取所有字段
		rows = nil
		sql = fmt.Sprintf("SELECT COLUMN_NAME from INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='%s' AND TABLE_NAME = '%s'", src.Cfg["dbname"], tableName)
		src.Db.Raw(sql).Scan(&rows)
		for _, row := range rows {
			src.Field = append(src.Field, row["COLUMN_NAME"].(string))
		}
	} else {
		src.FieldSql = my.GenerateFieldSql(src.Field)
	}
	// 构建唯一键sql
	if len(src.Key) == 0 {
		src.KeySql = my.GetUniqueKeySql(src.Db, src.Cfg["dbname"], tableName)
		if src.KeySql == "" && src.Count > int64(src.PageSize) {
			logr.L.Panic("表中没有唯一键，请在计划中配置Src的Key")
		}
	} else {
		src.KeySql = my.GenerateFieldSql(src.Key)
	}

	src.DataCacheMap = make([]map[string]interface{}, 0, src.PageSize)
	src.PrimaryValueList = make([]interface{}, 0, src.PageSize)
}

/*---------- mysql 目标 ----------*/

// 生成创建表的golang代码文件
func myGenerateCreateTableCode(dst *Dst, tableName []string) []string {
	code := fmt.Sprintf(`package main

import (
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/my"
	"go-com/internal/app"
	"go-com/runtime/dao/model"
)

func main() {
	config.Load()
	logr.InitLog("create_db_table")
	app.%s = my.NewDb(my.Config{Mysql: config.C.%s})

	migrator := app.%s.Migrator()

	var err error
`, dst.Cfg["db_app_name"], dst.Cfg["db_app_name"], dst.Cfg["db_app_name"])
	var err error
	var tableNeedCreate []string
	for _, t := range tableName {
		if strings.Contains(t, ".") {
			t = strings.Split(t, ".")[1]
		}

		if dst.Db.Migrator().HasTable(t) {
			continue
		}
		tableNeedCreate = append(tableNeedCreate, t)
		code += fmt.Sprintf("\tif err = migrator.CreateTable(&model.%s{}); err != nil {\n\t\tlogr.L.Panic(err)\n\t}\n", tool.SepNameToCamel(t))
	}
	code += "}"

	if len(tableNeedCreate) > 0 {
		if err = os.WriteFile(config.Root+"create_db_table.go", []byte(code), 0755); err != nil {
			logr.L.Panic(err)
		}
	}
	return tableNeedCreate
}
