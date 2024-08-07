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

// GenerateUpdFieldSql 将ON DUPLICATE KEY UPDATE字段列表转换成字段sql
func GenerateUpdFieldSql(fieldRaw []string) string {
	var updFieldSql string
	for _, field := range fieldRaw {
		updFieldSql += fmt.Sprintf("`%s`=VALUES(`%s`),", field, field)
	}
	updFieldSql = strings.TrimSuffix(updFieldSql, ",")
	return updFieldSql
}

// GetUniqueKeySql 查找指定表中的一个唯一索引列的sql呈现
func GetUniqueKeySql(db *gorm.DB, dbName string, tableName string) string {
	// 获取主键
	sql := fmt.Sprintf("SELECT COLUMN_NAME from INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA='%s' AND TABLE_NAME = '%s' AND COLUMN_KEY='PRI'", dbName, tableName)
	var rows []map[string]interface{}
	db.Raw(sql).Scan(&rows)
	var fields []string
	if len(rows) == 0 {
		// 获取唯一键
		rows = nil
		sql = fmt.Sprintf("SELECT CONSTRAINT_NAME FROM INFORMATION_SCHEMA.TABLE_CONSTRAINTS WHERE TABLE_SCHEMA='%s' AND TABLE_NAME = '%s'  and CONSTRAINT_TYPE='UNIQUE' LIMIT 1;", dbName, tableName)
		db.Raw(sql).Scan(&rows)
		if len(rows) == 0 {
			return ""
		}
		sql = fmt.Sprintf("SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA='%s' AND TABLE_NAME = '%s' AND INDEX_NAME='%s'", dbName, tableName, rows[0]["CONSTRAINT_NAME"])
		rows = nil
		db.Raw(sql).Scan(&rows)
	}

	for _, row := range rows {
		fields = append(fields, row["COLUMN_NAME"].(string))
	}
	return GenerateFieldSql(fields)
}

// GetDbTables 获取数据库的所有表
func GetDbTables(db *gorm.DB, dbname string) []map[string]interface{} {
	var rows []map[string]interface{}
	sql := fmt.Sprintf("SELECT TABLE_NAME as table_name FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA='%s' and TABLE_TYPE='BASE TABLE'", dbname)
	db.Raw(sql).Scan(&rows)
	return rows
}
