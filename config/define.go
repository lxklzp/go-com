package config

import (
	"database/sql/driver"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go-com/core/pg"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/http"
	"reflect"
	"strings"
	"time"
	"unicode"
)

const (
	EnvDev  = 1
	EnvProd = 2
)

const (
	DateTimeFormatter       = "2006-01-02 15:04:05"
	MonthNumberFormatter    = "200601"
	DateTimeNumberFormatter = "20060102150405"
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
var KafkaConsumeWorkerNumCh chan bool
var DelayQueueConsumeWorkerNumCh chan bool

var ServeApi *http.Server
var Pg *gorm.DB

func InitDefine() {
	KafkaConsumeWorkerNumCh = make(chan bool, C.App.MaxKafkaConsumeWorkerNum)
	DelayQueueConsumeWorkerNumCh = make(chan bool, C.App.MaxDelayQueueConsumeWorkerNum)
	tm, _ := time.ParseInLocation(DateTimeFormatter, "1980-01-01 00:00:00", time.Local)
	DefaultTimeMin = Timestamp(tm)
	tm, _ = time.ParseInLocation(DateTimeFormatter, "2034-01-01 00:00:00", time.Local)
	DefaultTimeMax = Timestamp(tm)

	Pg = pg.NewDb(C.Pg)
}

type Timestamp time.Time

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
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

func AddRouterApi(ctl interface{}) {
	ty := reflect.TypeOf(ctl)
	value := reflect.ValueOf(ctl)
	pathCtl := camelToSepName(ty.Name(), '-')
	numMethod := ty.NumMethod()
	for i := 0; i < numMethod; i++ {
		var routerApi RouterApi
		pathAction := camelToSepName(strings.TrimPrefix(ty.Method(i).Name, "Action"), '-')
		routerApi.Path = pathCtl + "/" + pathAction
		routerApi.Action = value.Method(i).Interface().(func(c *gin.Context) interface{})
		RouterApiList = append(RouterApiList, routerApi)
	}
}
