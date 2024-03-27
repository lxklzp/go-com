package my

import (
	"fmt"
	"go-com/config"
	"go-com/core/orm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

type Config struct {
	config.Mysql
}

func NewDb(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Addr, cfg.Dbname)
	return orm.NewDb(mysql.Open(dsn), orm.DbConfig{DbConfig: cfg.DbConfig})
}

// GenerateFieldSql 将字段列表转换成字段sql
func GenerateFieldSql(fieldRaw []string) string {
	var fieldSql string
	for _, field := range fieldRaw {
		fieldSql += fmt.Sprintf("`%s`,", field)
	}
	fieldSql = strings.TrimSuffix(fieldSql, ",")
	return fieldSql
}

// GetDbTables 获取数据库的所有表
func GetDbTables(db *gorm.DB, dbname string) []map[string]interface{} {
	var rows []map[string]interface{}
	sql := fmt.Sprintf("SELECT TABLE_NAME as table_name FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA='%s' and TABLE_TYPE='BASE TABLE'", dbname)
	db.Raw(sql).Scan(&rows)
	return rows
}
