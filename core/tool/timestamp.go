package tool

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"go-com/config"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Timestamp time.Time

func (t *Timestamp) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	str := string(data)
	timeStr := FormatFromTimeStandard(strings.Trim(str, "\""))
	tm, err := time.ParseInLocation(config.DateTimeFormatter, timeStr, time.Local)
	*t = Timestamp(tm)
	return err
}

func (t Timestamp) MarshalJSON() ([]byte, error) {
	tm := (time.Time)(t)
	if tm.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(fmt.Sprintf("\"%v\"", tm.Format(config.DateTimeFormatter))), nil
}

func (t Timestamp) Value() (driver.Value, error) {
	tm := (time.Time)(t)
	if tm.IsZero() {
		return nil, nil
	}
	return tm.Format(config.DateTimeFormatter), nil
}

func (t *Timestamp) Scan(v interface{}) error {
	switch vt := v.(type) {
	case time.Time:
		*t = Timestamp(vt)
	default:
		return errors.New("类型处理错误。")
	}
	return nil
}

func (t *Timestamp) String() string {
	tm := (*time.Time)(t)
	if tm.IsZero() {
		return ""
	}
	return tm.Format(config.DateTimeFormatter)
}

func (t *Timestamp) GobEncode() ([]byte, error) {
	tm := (*time.Time)(t)
	return tm.GobEncode()
}

func (t *Timestamp) GobDecode(data []byte) error {
	tm := (*time.Time)(t)
	return tm.GobDecode(data)
}

func (t Timestamp) GormDBDataType(db *gorm.DB) string {
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
