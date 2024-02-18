package global

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"go-com/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"unicode"
)

func init() {
}

func FormatStandardDatetime(datetime string) string {
	datetime = strings.Replace(datetime, "T", " ", 1)
	return datetime[0:19]
}

func CopyFile(dstName, srcName string) {
	src, err := os.Open(srcName)
	if err != nil {
		Log.Panic(err)
	}
	defer src.Close()

	dst, err := os.OpenFile(dstName, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		Log.Panic(err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		Log.Panic(err)
	}
}

// ErrorStack error返回错误栈信息
func ErrorStack(err interface{}) string {
	var msg string
	switch err.(type) {
	case error:
		msg = fmt.Sprintf("%+v", errors.WithStack(err.(error)))
	default:
		msg = fmt.Sprintf("%+v", err)
	}
	Log.Error(msg)
	return msg
}

type ResponseData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func RespData(code int, message string, data interface{}) ResponseData {
	return ResponseData{Code: code, Message: message, Data: data}
}

func CamelToSepName(field string, sep rune) string {
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

func SepNameToCamel(field string) string {
	return strings.ReplaceAll(cases.Title(language.English).String(strings.ReplaceAll(strings.ToLower(field), "_", " ")), " ", "")
}

func Get(url string, param map[string]string, header map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// header头
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// 参数
	query := req.URL.Query()
	for k, v := range param {
		query.Add(k, v)
	}
	req.URL.RawQuery = query.Encode()

	// 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	// 处理返回结果
	result, err := io.ReadAll(resp.Body)
	if config.C.App.DebugMode {
		Log.Debugf("请求 %s:%s\n响应 [%d] %s", url, param, resp.StatusCode, result)
	}
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New("服务器异常，响应码：" + strconv.Itoa(resp.StatusCode))
	} else {
		return result, nil
	}
}

func Post(url string, param []byte, header map[string]string) ([]byte, error) {
	Log.Info(string(param))
	body := bytes.NewBuffer(param)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	// header头
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}

	// 请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	// 处理返回结果
	result, err := io.ReadAll(resp.Body)
	if config.C.App.DebugMode {
		Log.Debugf("请求 %s:%s\n响应 [%d] %s", url, param, resp.StatusCode, result)
	}
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, errors.New("服务器异常，响应码：" + strconv.Itoa(resp.StatusCode))
	} else {
		return result, nil
	}
}

// InArray 值是否在切片中
func InArray[T int | string](value T, array []T) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}
	return false
}

// NarrowString 半角转全角
func NarrowString(str string) string {
	list := []string{"，", ",", "《", "<", "。", ".", "》", ">", "？", "?", "；", ";", "：", ":", "‘", "'", "”", "\"", "【", "[", "】", "]", "！", "!", "（", "(", "）", ")", "——", "-"}
	replace := strings.NewReplacer(list...)
	return replace.Replace(str)
}

// InterfaceToString interface转string
func InterfaceToString(v interface{}) string {
	if v == nil {
		return ""
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
	default:
		return ""
	}
}

// JsonInterfaceToString json解析的interface转string
func JsonInterfaceToString(value interface{}) string {
	switch value.(type) {
	case string:
		return value.(string)
	case float64:
		return strconv.FormatFloat(value.(float64), 'f', 6, 64)
	}
	return ""
}

// ExitNotify 监听退出信号，关闭系统资源
func ExitNotify(close func()) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		for s := range ch {
			switch s {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
				Log.Debug("关闭系统")
				close()
				os.Exit(0)
			}
		}
	}()
}

// IPString2Long 把ip字符串转为数值
func IPString2Long(ip string) (uint, error) {
	b := net.ParseIP(ip).To4()
	if b == nil {
		return 0, errors.New("invalid ipv4 format")
	}

	return uint(b[3]) | uint(b[2])<<8 | uint(b[1])<<16 | uint(b[0])<<24, nil
}

// Long2IPString 把数值转为ip字符串
func Long2IPString(i uint) (string, error) {
	if i > math.MaxUint32 {
		return "", errors.New("beyond the scope of ipv4")
	}

	ip := make(net.IP, net.IPv4len)
	ip[0] = byte(i >> 24)
	ip[1] = byte(i >> 16)
	ip[2] = byte(i >> 8)
	ip[3] = byte(i)

	return ip.String(), nil
}
