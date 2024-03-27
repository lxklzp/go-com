package pg

import (
	"fmt"
	"go-com/config"
	"go-com/core/orm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
)

type Config struct {
	config.Postgresql
}

// NewDb 实例化gorm的postgresql连接
func NewDb(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", cfg.Host, cfg.User, cfg.Password, cfg.Dbname, cfg.Port)
	return orm.NewDb(postgres.Open(dsn), orm.DbConfig{DbConfig: cfg.DbConfig})
}

// GenerateFieldSql 将字段列表转换成字段sql
func GenerateFieldSql(fieldRaw []string) string {
	var fieldSql string
	for _, field := range fieldRaw {
		fieldSql += fmt.Sprintf(`"%s",`, field)
	}
	fieldSql = strings.TrimSuffix(fieldSql, ",")
	return fieldSql
}

// GetSchemaTableName 从完整表名中提取schema和table的名称
func GetSchemaTableName(tableName string) (string, string) {
	schemaName := "public"
	names := strings.Split(tableName, ".")
	if len(names) == 2 {
		schemaName = names[0]
		tableName = names[1]
	}
	return schemaName, tableName
}

// GetDbTables 获取数据库的所有表
func GetDbTables(db *gorm.DB, schema string) []map[string]interface{} {
	var rows []map[string]interface{}
	sql := fmt.Sprintf("select tablename from pg_tables WHERE schemaname='%s'", schema)
	db.Raw(sql).Scan(&rows)
	return rows
}
