package config

import (
	"bytes"
	"database/sql/driver"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"reflect"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	EnvDev  = 1
	EnvProd = 2
)

const (
	Version                   = "v1.0.1"
	DateTimeFormatter         = "2006-01-02 15:04:05"
	DateTimeStandardFormatter = "2006-01-02T15:04:05"
	DateFormatter             = "2006-01-02"
	TimeFormatter             = "15:04:05"
	MonthNumberFormatter      = "200601"
	DateTimeNumberFormatter   = "20060102150405"
	DateNumberFormatter       = "20060102"

	TypeEnable  = 1
	TypeDisable = 2

	Sep      = "?#"
	MinFloat = float64(-9007199254740992) // -2^53
	MaxFloat = float64(9007199254740992)  // 2^53
)

const (
	_        = iota
	KB int64 = 1 << (10 * iota)
	MB
	GB
	TB
)

var DefaultTimeMin Timestamp
var DefaultTimeMax Timestamp

var Trans ut.Translator

var BufPool *sync.Pool // bytes.Buffer 缓存池

func InitDefine() {
	tm, _ := time.ParseInLocation(DateTimeFormatter, "1980-01-01 00:00:00", time.Local)
	DefaultTimeMin = Timestamp(tm)
	tm, _ = time.ParseInLocation(DateTimeFormatter, "2034-01-01 00:00:00", time.Local)
	DefaultTimeMax = Timestamp(tm)

	Trans, _ = ut.New(en.New(), zh.New()).GetTranslator("zh")

	BufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4096))
		},
	}
}

type Timestamp time.Time

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	var err error
	str := string(data)
	timeStr := strings.Trim(str, "\"")
	t1, err := time.Parse(DateTimeFormatter, timeStr)
	*t = Timestamp(t1)
	return err
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", time.Time(t).Format(DateTimeFormatter))), nil
}

func (t Timestamp) Value() (driver.Value, error) {
	return time.Time(t).Format(DateTimeFormatter), nil
}

func (t *Timestamp) Scan(v interface{}) error {
	switch vt := v.(type) {
	case time.Time:
		*t = Timestamp(vt)
	default:
		return errors.New("类型处理错误")
	}
	return nil
}

func (t *Timestamp) String() string {
	return time.Time(*t).Format(DateTimeFormatter)
}

func (Timestamp) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "mysql":
		return "datetime"
	case "postgres":
		return "timestamptz"
	case "sqlite":
		return "TEXT"
	default:
		return ""
	}
}

type RouterApi struct {
	Path   string
	Action func(c *gin.Context) interface{}
}

var RouterApiList []RouterApi

func camelToSepName(field string, sep rune) string {
	var buffer []rune
	for i, r := range []rune(field) {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer = append(buffer, sep)
			}
			buffer = append(buffer, unicode.ToLower(r))
		} else {
			buffer = append(buffer, r)
		}
	}
	return string(buffer)
}

func AddRouterApi(ctl interface{}, routerApiList *[]RouterApi) {
	ty := reflect.TypeOf(ctl)
	value := reflect.ValueOf(ctl)
	pathCtl := camelToSepName(ty.Name(), '-')
	numMethod := ty.NumMethod()
	for i := 0; i < numMethod; i++ {
		var routerApi RouterApi
		methodName := ty.Method(i).Name
		if strings.HasPrefix(methodName, "Action") {
			pathAction := camelToSepName(strings.TrimPrefix(methodName, "Action"), '-')
			routerApi.Path = C.App.Prefix + "/" + pathCtl + "/" + pathAction
			routerApi.Action = value.Method(i).Interface().(func(c *gin.Context) interface{})
			*routerApiList = append(*routerApiList, routerApi)
		}
	}
}
