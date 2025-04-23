package tool

import (
	"fmt"
	"github.com/pkg/errors"
	"regexp"
	"strconv"
)

var regFloat, regInt *regexp.Regexp

func init() {
	regFloat, _ = regexp.Compile(`^[\-\d.]+`)
	regInt, _ = regexp.Compile(`^[\-\d]+`)
}

// InterfaceToString interface转string
func InterfaceToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch vt := value.(type) {
	case string:
		return vt
	case []byte:
		return string(vt)
	case bool:
		return strconv.FormatBool(vt)
	case error:
		return vt.Error()
	case fmt.Stringer:
		return vt.String()
	case float32:
		return strconv.FormatFloat(float64(vt), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(vt, 'f', -1, 64)
	case int:
		return strconv.Itoa(vt)
	case int8:
		return strconv.Itoa(int(vt))
	case int16:
		return strconv.Itoa(int(vt))
	case int32:
		return strconv.Itoa(int(vt))
	case int64:
		return strconv.FormatInt(vt, 10)
	case uint:
		return strconv.FormatUint(uint64(vt), 10)
	case uint8:
		return strconv.FormatUint(uint64(vt), 10)
	case uint16:
		return strconv.FormatUint(uint64(vt), 10)
	case uint32:
		return strconv.FormatUint(uint64(vt), 10)
	case uint64:
		return strconv.FormatUint(vt, 10)
	default:
		return fmt.Sprint(value)
	}
}

// InterfaceToFloat interface转float
func InterfaceToFloat(value interface{}) (float64, error) {
	if value == nil {
		return 0, errors.New("不支持的类型。")
	}

	switch vt := value.(type) {
	case string:
		return strconv.ParseFloat(regFloat.FindString(vt), 10)
	case float64:
		return vt, nil
	case float32:
		return float64(vt), nil
	case int:
		return float64(vt), nil
	case int8:
		return float64(vt), nil
	case int16:
		return float64(vt), nil
	case int32:
		return float64(vt), nil
	case int64:
		return float64(vt), nil
	case uint:
		return float64(vt), nil
	case uint8:
		return float64(vt), nil
	case uint16:
		return float64(vt), nil
	case uint32:
		return float64(vt), nil
	case uint64:
		return float64(vt), nil
	default:
		return 0, errors.New("不支持的类型。")
	}
}

// InterfaceToInt interface转int
func InterfaceToInt(value interface{}) (int, error) {
	if value == nil {
		return 0, errors.New("不支持的类型。")
	}

	switch vt := value.(type) {
	case string:
		return strconv.Atoi(regInt.FindString(vt))
	case float64:
		return int(vt), nil
	case float32:
		return int(vt), nil
	case int:
		return vt, nil
	case int8:
		return int(vt), nil
	case int16:
		return int(vt), nil
	case int32:
		return int(vt), nil
	case int64:
		return int(vt), nil
	case uint:
		return int(vt), nil
	case uint8:
		return int(vt), nil
	case uint16:
		return int(vt), nil
	case uint32:
		return int(vt), nil
	case uint64:
		return int(vt), nil
	default:
		return 0, errors.New("不支持的类型。")
	}
}
