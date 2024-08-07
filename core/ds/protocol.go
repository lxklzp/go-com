package ds

import (
	"go-com/config"
	"go-com/core/tool"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"time"
)

const (
	ForDataBase          = "__DATABASE" // table为这个值时，同步全库的表
	NullStr              = "__NULL__"
	DefaultValueDatetime = "__DATETIME"
)

const (
	DataTypeMy            = "my"
	DataTypePg            = "pg"
	DataTypeOra           = "ora"
	SyncReadScrollTimeout = 60 * 20 // scroll_id缓存时间，20分钟
)

var dataSyncReadCache sync.Map // 分页同步时的缓存状态数据，用于接口同步数据等

// interface转string
func interfaceToString(v interface{}) string {
	if v == nil {
		return NullStr
	}
	switch v.(type) {
	case string:
		return v.(string)
	case float64:
		return strconv.FormatFloat(v.(float64), 'f', -1, 64)
	case []byte:
		return string(v.([]byte))
	case int:
		return strconv.Itoa(v.(int))
	case int8:
		return strconv.Itoa(int(v.(int8)))
	case int16:
		return strconv.Itoa(int(v.(int16)))
	case int32:
		return strconv.Itoa(int(v.(int32)))
	case int64:
		return strconv.FormatInt(v.(int64), 10)
	case uint:
		return strconv.FormatUint(uint64(v.(uint)), 10)
	case uint8:
		return strconv.FormatUint(uint64(v.(uint8)), 10)
	case uint16:
		return strconv.FormatUint(uint64(v.(uint16)), 10)
	case uint32:
		return strconv.FormatUint(uint64(v.(uint32)), 10)
	case uint64:
		return strconv.FormatUint(v.(uint64), 10)
	case time.Time:
		return v.(time.Time).Format(config.DateTimeFormatter)
	default:
		return NullStr
	}
}

// 格式化默认值
func formatDefaultValue(flag string) string {
	switch flag {
	case DefaultValueDatetime:
		return time.Now().Format(config.DateTimeFormatter)
	}
	return flag
}

// Src 数据源
type Src struct {
	// plan配置项
	Cfg         map[string]string // 数据库连接配置
	Table       interface{}       // 单表：字符串，单表可以精细化操作；多表：字符串数组，同步列举的所有表，以及表内所有字段；全表：值为__DATABASE时，同步数据库所有表，以及表内所有字段
	Key         []string          // 单表：唯一键，用于分页排序；为空时的情况：pgsql自动查询唯一索引，如果没有，并且总条数超过每次读取条数，报错。
	Field       []string          // 单表：同步字段，为空时同步所有字段
	Where       string            // 单表：查询条件
	PageSize    int               // 每次读取条数，分页的每页条数，源分页应为目标分页的整数
	Column      []int             // 需要同步的列号，excel使用
	CsvFilename string            // 读取到csv文件的文件全路径
	Schema      string            // postgresql全库创建表时使用/excel的sheet名称

	// 运行时属性
	Db               *gorm.DB                 // 数据库连接
	Page             int                      // 分页的当前页数
	Count            int64                    // 总条数
	DataCacheMap     []map[string]interface{} // 缓存读取的数据，map元素
	DataCacheArray   [][]string               // 缓存读取的数据，slice元素
	PrimaryValueList []interface{}            // 主键值列表
	FieldSql         string                   // 同步字段sql
	KeySql           string                   // 唯一键sql
}

// Dst 数据目标
type Dst struct {
	// plan配置项
	Cfg          map[string]string // 数据库连接配置
	Table        string            // 单表：表名
	Field        []string          // 单表：同步字段
	UpdField     []string          // 单表：mysql数据存在时，更新字段
	DefaultField []string          // 单表：默认字段
	DefaultValue []interface{}     // 单表：默认字段值，__DATETIME表示当前时间的2024-01-18 17:09:36格式
	PageSize     int               // 每次读取条数，分页的每页条数，源分页应为目标分页的整数
	Truncate     bool              // 写入前是否清空表
	EndSql       string            // 写入完成后执行的sql
	Schema       string            // postgresql全库创建表时使用/csv的列分隔符

	// 运行时属性
	Db          *gorm.DB       // 数据库连接
	Step        int64          // 已写入的条数
	StepPeriod  int64          // 在一轮中已写入的条数
	FieldSql    string         // 同步字段sql
	UpdFieldSql string         // 数据存在时，更新字段sql
	ValueSql    []byte         // 一行值的sql拼接缓存
	Sql         []byte         // sql缓存
	SlicePage   tool.SlicePage // 切片分页工具
}

// DataSync 同步组件
type DataSync struct {
	Src         Src
	Dst         Dst
	DataTypeSrc string
	DataTypeDst string
}

type ReadScrollData struct {
	ScrollId string                   `json:"scroll_id"`
	List     []map[string]interface{} `json:"list"`
	Count    int64                    `json:"count"`
}
