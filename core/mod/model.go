package mod

import (
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"go-com/config"
	"go-com/core/logr"
	"gorm.io/gorm"
	"math"
	"os"
	"strings"
	"time"
)

const MaxPageRead = 5000
const MaxPageWrite = 200

type Base struct {
	TimeFrom string `gorm:"-" json:"time_from"`
	TimeTo   string `gorm:"-" json:"time_to"`
	Page     int    `gorm:"-" json:"pageNo"`
	PageSize int    `gorm:"-" json:"pageSize"`
}

type PrimaryId struct {
	ID int64 `json:"id"`
}

type PrimaryIdName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Name struct {
	Name string `json:"name"`
}

type Trim string

func (t *Trim) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*t = Trim(strings.TrimSpace(s))
	return nil
}

func (b *Base) Validate() {
	if b.Page == 0 {
		b.Page = 1
	}
	if b.PageSize == 0 {
		b.PageSize = 20
	} else if b.PageSize > MaxPageRead {
		b.PageSize = MaxPageRead
	}
}

func IsZero(value interface{}) bool {
	switch value.(type) {
	case string:
		if value == "" {
			return true
		}
	case Trim:
		if value == Trim("") || value == Trim("%%") {
			return true
		}
	case int8:
		if value == int8(0) {
			return true
		}
	case int32:
		if value == int32(0) {
			return true
		}
	case int64:
		if value == int64(0) {
			return true
		}
	case int:
		if value == 0 {
			return true
		}
	case []int8:
		if len(value.([]int8)) == 0 {
			return true
		}
	case []int32:
		if len(value.([]int32)) == 0 {
			return true
		}
	case []int64:
		if len(value.([]int64)) == 0 {
			return true
		}
	case []int:
		if len(value.([]int)) == 0 {
			return true
		}
	case config.Timestamp:
		if time.Time(value.(config.Timestamp)).IsZero() {
			return true
		}
	}
	return false
}

func FilterWhere(db *gorm.DB, query interface{}, arg interface{}) {
	if !IsZero(arg) {
		db.Where(query, arg)
	}
}

// ExcelReadTable 从数据库中查询的数据格式
type ExcelReadTable struct {
	Count int64
	List  interface{}
}

// ExcelExport 导出成excel
func ExcelExport(name string, title []interface{}, readTable func(page int, pageSize int, isCount bool) ExcelReadTable, streamWrite func(stream *excelize.StreamWriter, table ExcelReadTable, rowNext *int)) (string, error) {
	// 初始化excel和写入流
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			logr.L.Error(err)
		}
	}()
	stream, err := f.NewStreamWriter("Sheet1")
	if err != nil {
		return "", err
	}

	// 写入标题
	if err = stream.SetRow("A1", title); err != nil {
		return "", err
	}
	rowNext := 2
	var table ExcelReadTable
	// 分页 从数据库中查询第一页数据
	table = readTable(1, MaxPageRead, true)

	// 第一页有数据
	if table.Count != int64(0) {
		// 将第一页数据写入excel
		streamWrite(stream, table, &rowNext)
		// 总页数
		pageCount := int(math.Ceil(float64(table.Count) / float64(MaxPageRead)))
		if pageCount > 1 {
			// 分页读取数据库表并写入excel
			for i := 2; i <= pageCount; i++ {
				table = readTable(i, MaxPageRead, false)
				streamWrite(stream, table, &rowNext)
			}
		}
	}
	if err = stream.Flush(); err != nil {
		return "", err
	}

	// 创建导出文件夹
	path := config.RuntimePath + "/public"
	relativePath := "/" + time.Now().Format(config.MonthNumberFormatter)
	path += relativePath
	if err = os.MkdirAll(path, 0755); err != nil {
		return "", err
	}
	filename := fmt.Sprintf("/%s_%s.xlsx", name, time.Now().Format(config.DateTimeNumberFormatter))
	path += filename
	relativePath += filename

	if err = f.SaveAs(path); err != nil {
		return "", err
	}

	return "/export" + relativePath, nil
}

// SlicePage 切片分页
type SlicePage struct {
	totalCount int64
	page       int64
	pageSize   int64
	From       int64
	To         int64
}

func NewSlicePage(totalCount int64, pageSize int64) *SlicePage {
	var sp SlicePage
	sp.totalCount = totalCount
	sp.pageSize = pageSize
	return &sp
}

func (sp *SlicePage) Next() bool {
	if sp.To >= sp.totalCount {
		return false
	}
	sp.From = sp.page * sp.pageSize
	sp.page++
	if sp.totalCount <= sp.page*sp.pageSize {
		sp.To = sp.totalCount
	} else {
		sp.To = sp.page * sp.pageSize
	}
	return true
}
