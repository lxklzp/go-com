package orm

import (
	"fmt"
	"go-com/config"
	"gorm.io/gorm"
	"math"
	"strconv"
	"strings"
)

const (
	DbMysql = 1
	DbPgsql = 2

	TypeInt    = "int"
	TypeFloat  = "float"
	TypeString = "string"
)

var MethodAdapter methodAdapter

type methodAdapter struct {
}

// FindInSet 实现FindInSet
func (ma methodAdapter) FindInSet(value interface{}, field string) string {
	switch config.C.App.DbType {
	case DbMysql:
		return fmt.Sprintf("FIND_IN_SET('%v',%s)", value, field)
	case DbPgsql:
		return fmt.Sprintf("'%v' = ANY (STRING_TO_ARRAY(%s, ','))", value, field)
	}
	return ""
}

// GetTableNameFull 获取完整表名，处理pgsql表名问题
func (ma methodAdapter) GetTableNameFull(schema string, tableName string) string {
	switch config.C.App.DbType {
	case DbMysql:
		return tableName
	case DbPgsql:
		return fmt.Sprintf(`"%s"."%s"`, schema, tableName)
	}
	return ""
}

// CreateTableLike 复制表结构
func (ma methodAdapter) CreateTableLike(db *gorm.DB, schema string, newTableName string, currentTableName string) error {
	var sql string
	switch config.C.App.DbType {
	case DbMysql:
		sql = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s LIKE %s`, newTableName, currentTableName)
	case DbPgsql:
		newTableName = ma.GetTableNameFull(schema, newTableName)
		currentTableName = ma.GetTableNameFull(schema, currentTableName)
		sql = fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s ( LIKE %s INCLUDING COMMENTS INCLUDING DEFAULTS INCLUDING INDEXES INCLUDING CONSTRAINTS )`, newTableName, currentTableName)
	}
	return db.Exec(sql).Error
}

type TableColumn struct {
	ColumnName string `gorm:"column:column_name" json:"column_name"`
	DataType   string `gorm:"column:data_type" json:"data_type"`
}

// GetTableColumnSql 获取表字段sql
func (ma methodAdapter) GetTableColumnSql(schema string, tableName string) string {
	return fmt.Sprintf("SELECT column_name,data_type FROM information_schema.columns WHERE table_schema='%s' AND table_name='%s' ORDER BY ordinal_position ASC", schema, tableName)
}

// FormatTableColumnDataType 格式化表字段数据结构
func (ma methodAdapter) FormatTableColumnDataType(list []TableColumn) {
	for k := range list {
		if strings.Contains(list[k].DataType, "int") {
			list[k].DataType = TypeInt
		} else if strings.Contains(list[k].DataType, "decimal") || strings.Contains(list[k].DataType, "double") || strings.Contains(list[k].DataType, "float") || strings.Contains(list[k].DataType, "numeric") {
			list[k].DataType = TypeFloat
		} else {
			list[k].DataType = TypeString
		}
	}
}

// GetTableColumn 获取表字段
func (ma methodAdapter) GetTableColumn(db *gorm.DB, schema string, tableName string) ([]TableColumn, error) {
	var list []TableColumn
	err := db.Raw(ma.GetTableColumnSql(schema, tableName)).Scan(&list).Error
	if err != nil {
		return nil, err
	}
	ma.FormatTableColumnDataType(list)
	return list, nil
}

// FormatStringValue 格式化字符串类型的值
func (ma methodAdapter) FormatStringValue(ty string, v string) interface{} {
	switch ty {
	case TypeInt:
		value, _ := strconv.Atoi(v)
		return value
	case TypeFloat:
		value, _ := strconv.ParseFloat(v, 64)
		if math.IsNaN(value) {
			return 0
		} else {
			return value
		}
	case TypeString:
		return v
	}
	return ""
}
