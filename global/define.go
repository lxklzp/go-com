package global

import "C"
import (
	"database/sql/driver"
	"fmt"
	"github.com/pkg/errors"
	"go-com/config"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"net/http"
	"strings"
	"time"
)

var ServeApi *http.Server

const (
	DateTimeFormatter         = "2006-01-02 15:04:05"
	DateHourNumberFormatter   = "2006010215"
	DateTimeStandardFormatter = "2006-01-02T15:04:05Z"
	DateTimeFormatterNumber   = "20060102150405"
	MonthFormatter            = "200601"
	LetterBytes               = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var DefaultTimeMin Timestamp
var DefaultTimeMax Timestamp
var KafkaConsumeWorkerNumCh chan bool
var DelayQueueConsumeWorkerNumCh chan bool

func InitDefine() {
	KafkaConsumeWorkerNumCh = make(chan bool, config.C.App.MaxKafkaConsumeWorkerNum)
	DelayQueueConsumeWorkerNumCh = make(chan bool, config.C.App.MaxDelayQueueConsumeWorkerNum)
	tm, _ := time.ParseInLocation(DateTimeFormatter, "1980-01-01 00:00:00", time.Local)
	DefaultTimeMin = Timestamp(tm)
	tm, _ = time.ParseInLocation(DateTimeFormatter, "2034-01-01 00:00:00", time.Local)
	DefaultTimeMax = Timestamp(tm)
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
