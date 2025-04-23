package protocol

import (
	"go-com/config"
	"go-com/core/orm"
	"go-com/internal/app"
)

// 约定
const (
	TypeEnable  = 1
	TypeDisable = 2
)

func GetDbSchema() string {
	switch config.C.App.DbType {
	case orm.DbMysql:
		return config.C.Mysql.Dbname
	case orm.DbPgsql:
		return config.C.Pg.Schema
	}
	return ""
}

func GetTableColumn(tableName string) ([]orm.TableColumn, error) {
	return orm.MethodAdapter.GetTableColumn(app.Db, GetDbSchema(), tableName)
}

func GetTableNameFull(tableName string) string {
	return orm.MethodAdapter.GetTableNameFull(GetDbSchema(), tableName)
}
