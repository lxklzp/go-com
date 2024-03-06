package pg

import (
	"fmt"
	"go-com/core/logr"
	"go-com/core/orm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
	orm.DbConfig
}

// NewDb 实例化gorm的postgresql连接
func NewDb(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable", cfg.Host, cfg.User, cfg.Password, cfg.Dbname, cfg.Port)
	return orm.NewDb(postgres.Open(dsn), cfg.DbConfig)
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

// GetUniqueKeySql 查找指定表中的一个唯一索引列的sql呈现
func GetUniqueKeySql(db *gorm.DB, tableName string) string {
	schemaName, tableName := GetSchemaTableName(tableName)
	var rows []map[string]interface{}
	sql := fmt.Sprintf("select indexdef from pg_indexes where schemaname='%s' and tablename='%s';", schemaName, tableName)
	db.Raw(sql).Scan(&rows)
	for _, row := range rows {
		indexdef := row["indexdef"].(string)
		if strings.Contains(indexdef, "UNIQUE") {
			reg, _ := regexp.Compile(`\((.+)\)`)
			return reg.FindStringSubmatch(indexdef)[1]
		}
	}
	return ""
}

// Gentool 执行gentool指令
func Gentool(dbCfg map[string]string, tableName string) {
	var err error
	var res []byte

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbCfg["host"], dbCfg["user"], dbCfg["password"], dbCfg["dbname"], dbCfg["port"])
	if runtime.GOOS == "windows" {
		res, err = exec.Command("gentool", "-db", "postgres", "-dsn", dsn, "-onlyModel", "-tables", tableName).CombinedOutput()
	} else {
		res, err = exec.Command("sh", "-c", fmt.Sprintf("gentool -db postgres -dsn %s sslmode=disable -onlyModel -tables %s", dsn, tableName)).CombinedOutput()
	}
	if err != nil {
		logr.L.Fatal(err)
	}
	logr.L.Info(string(res))
}

// GetDbTables 获取数据库的所有表
func GetDbTables(db *gorm.DB, schema string) []map[string]interface{} {
	var rows []map[string]interface{}
	sql := fmt.Sprintf("select tablename from pg_tables WHERE schemaname='%s'", schema)
	db.Raw(sql).Scan(&rows)
	return rows
}
