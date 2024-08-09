package ora

import (
	"fmt"
	oracle "github.com/godoes/gorm-oracle"
	"go-com/config"
	"go-com/core/orm"
	"gorm.io/gorm"
	"strings"
)

type Config struct {
	config.Oracle
}

// NewDb 实例化gorm的postgresql连接
func NewDb(cfg Config) *gorm.DB {
	db := orm.NewDb(oracle.New(oracle.Config{
		DSN:                     oracle.BuildUrl(cfg.Host, cfg.Port, cfg.Service, cfg.User, cfg.Password, nil),
		IgnoreCase:              false, // query conditions are not case-sensitive
		NamingCaseSensitive:     true,  // whether naming is case-sensitive
		VarcharSizeIsCharLength: true,  // whether VARCHAR type size is character length, defaulting to byte length
	}), orm.DbConfig{DbConfig: cfg.DbConfig})

	if sqlDB, err := db.DB(); err == nil {
		_, _ = oracle.AddSessionParams(sqlDB, map[string]string{
			"TIME_ZONE":               "+08:00",                       // ALTER SESSION SET TIME_ZONE = '+08:00';
			"NLS_DATE_FORMAT":         "YYYY-MM-DD",                   // ALTER SESSION SET NLS_DATE_FORMAT = 'YYYY-MM-DD';
			"NLS_TIME_FORMAT":         "HH24:MI:SSXFF",                // ALTER SESSION SET NLS_TIME_FORMAT = 'HH24:MI:SS.FF3';
			"NLS_TIMESTAMP_FORMAT":    "YYYY-MM-DD HH24:MI:SSXFF",     // ALTER SESSION SET NLS_TIMESTAMP_FORMAT = 'YYYY-MM-DD HH24:MI:SS.FF3';
			"NLS_TIME_TZ_FORMAT":      "HH24:MI:SS.FF TZR",            // ALTER SESSION SET NLS_TIME_TZ_FORMAT = 'HH24:MI:SS.FF3 TZR';
			"NLS_TIMESTAMP_TZ_FORMAT": "YYYY-MM-DD HH24:MI:SSXFF TZR", // ALTER SESSION SET NLS_TIMESTAMP_TZ_FORMAT = 'YYYY-MM-DD HH24:MI:SS.FF3 TZR';
		})
	}
	return db
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

// GetUniqueKeySql 查找指定表中的一个唯一索引列的sql呈现
func GetUniqueKeySql(db *gorm.DB, username string, tableName string) string {
	// 获取主键
	sql := fmt.Sprintf("SELECT cu.COLUMN_NAME FROM user_cons_columns cu,user_constraints au WHERE cu.constraint_name = au.constraint_name AND cu.owner='%s' AND au.constraint_type = 'P' AND au.table_name = '%s'", username, tableName)
	var rows []map[string]interface{}
	db.Raw(sql).Scan(&rows)
	var fields []string
	if len(rows) == 0 {
		// 获取唯一键
		rows = nil
		sql = fmt.Sprintf("SELECT INDEX_NAME FROM user_indexes WHERE uniqueness ='UNIQUE' AND table_owner='%s' AND table_name='%s' AND rownum<=1", username, tableName)
		db.Raw(sql).Scan(&rows)
		if len(rows) == 0 {
			return ""
		}
		sql = fmt.Sprintf("SELECT COLUMN_NAME FROM all_ind_columns WHERE table_owner='%s' AND table_name='%s' AND index_name = '%s'", username, tableName, rows[0]["INDEX_NAME"])
		rows = nil
		db.Raw(sql).Scan(&rows)
	}

	for _, row := range rows {
		fields = append(fields, row["COLUMN_NAME"].(string))
	}
	return GenerateFieldSql(fields)
}

// GetDbTables 获取数据库的所有表
func GetDbTables(db *gorm.DB) []map[string]interface{} {
	var rows []map[string]interface{}
	sql := `SELECT TABLE_NAME as "table_name" FROM user_tables`
	db.Raw(sql).Scan(&rows)
	return rows
}
