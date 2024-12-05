package ds

import (
	"fmt"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/pg"
	"go-com/core/tool"
	"os"
	"strings"
)

/*---------- postgresql 源 ----------*/

// 读取表数据前的准备工作
func pgBeforeReadSingleTable(src *Src, tableName string) {
	tableName = strings.Trim(strings.Split(tableName, ".")[1], `"`)
	var err error
	var sql string
	var rows []map[string]interface{}
	if src.Where == "" {
		// 查询表数据总数，预估值
		sql = "SELECT reltuples::bigint FROM pg_catalog.pg_class WHERE relname = '" + tableName + "'"
		src.Db.Raw(sql).Scan(&rows)
		if len(rows) == 0 {
			logr.L.Panicf("源库中%s表不存在", tableName)
		}
		src.Count = rows[0]["reltuples"].(int64)
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
		sql = fmt.Sprintf("select column_name from information_schema.columns where table_schema='%s' and table_name='%s' order by ordinal_position", src.Schema, tableName)
		src.Db.Raw(sql).Scan(&rows)
		for _, row := range rows {
			src.Field = append(src.Field, row["column_name"].(string))
		}
	} else {
		src.FieldSql = pg.GenerateFieldSql(src.Field)
	}
	// 构建唯一键sql
	if len(src.Key) == 0 {
		src.KeySql = pg.GetUniqueKeySql(src.Db, src.Schema, tableName)
		if src.KeySql == "" && src.Count > int64(src.PageSize) {
			logr.L.Panic("表中没有唯一键，请在计划中配置Src的Key")
		}
	} else {
		src.KeySql = pg.GenerateFieldSql(src.Key)
	}

	src.DataCacheMap = make([]map[string]interface{}, 0, src.PageSize)
	src.PrimaryValueList = make([]interface{}, 0, src.PageSize)
}

/*---------- postgresql 目标 ----------*/

// 生成创建表的golang代码文件
func pgGenerateCreateTableCode(dst *Dst, tableName string) string {
	code := `package main

import (
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/orm"
	"go-com/core/pg"
	"go-com/dao/model"
)

func main() {
	config.Load()
	config.C.App.Id = 999
	logr.InitLog("create_table")

	var err error
`
	code += fmt.Sprintf("\tdb := pg.NewDb(pg.Config{Host: \"%s\", Port: \"%s\", User: \"%s\", Password: \"%s\", Dbname: \"%s\", DbConfig: orm.DbConfig{}})\n\tmigrator := db.Migrator()\n",
		dst.Cfg["host"],
		dst.Cfg["port"],
		dst.Cfg["user"],
		dst.Cfg["password"],
		dst.Cfg["dbname"])
	var err error
	var sql string
	rows := make([]map[string]interface{}, 0, 1024)
	var tableNeedCreate string
	for _, t := range strings.Split(tableName, ",") {
		sql = fmt.Sprintf("select tablename from pg_tables WHERE schemaname='%s' and tablename='%s'", dst.Schema, t)
		rows = nil
		dst.Db.Raw(sql).Scan(&rows)
		// 表不存在则创建
		if len(rows) == 0 {
			tableNeedCreate += t + ","
			code += fmt.Sprintf("\tif err = migrator.CreateTable(&model.%s{}); err != nil {\n\t\tlogr.L.Panic(err)\n\t}\n", tool.SepNameToCamel(t, true))
		}
	}
	code += "}"

	if tableNeedCreate != "" {
		if err = os.WriteFile(config.Root+"create_table.go", []byte(code), 0755); err != nil {
			logr.L.Panic(err)
		}
	}
	return strings.TrimRight(tableNeedCreate, ",")
}
