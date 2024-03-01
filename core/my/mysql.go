package my

import (
	"fmt"
	"go-com/core/orm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
)

type Config struct {
	Addr     string
	User     string
	Password string
	Dbname   string
	orm.DbConfig
}

func NewDb(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Addr, cfg.Dbname)
	return orm.NewDb(mysql.Open(dsn), cfg.DbConfig)
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
