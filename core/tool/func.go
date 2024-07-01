package tool

import (
	"bytes"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go-com/config"
	"go-com/core/logr"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// ExitNotify 监听退出信号，关闭系统资源
func ExitNotify(close func()) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		for s := range ch {
			switch s {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
				logr.L.Info("关闭系统")
				close()
				os.Exit(0)
			}
		}
	}()
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

// ErrorStack error返回错误栈信息
func ErrorStack(err interface{}) string {
	var msg string
	switch err.(type) {
	case error:
		msg = fmt.Sprintf("%+v", errors.WithStack(err.(error)))
	default:
		msg = fmt.Sprintf("%+v", err)
	}

	logr.L.Error(msg)
	return msg
}

// ErrorStr 增加表单验证错误信息的中文翻译
func ErrorStr(err error) string {
	switch err.(type) {
	case validator.ValidationErrors:
		return fmt.Sprintf("%s", err.(validator.ValidationErrors).Translate(config.Trans))
	}
	return err.Error()
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

func httpReqResp(req *http.Request, url string, param interface{}) ([]byte, error) {
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
	logr.L.Debugf("请求 %s:%s\n响应 [%d] %s", url, param, resp.StatusCode, result)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return result, errors.New("服务器异常，响应码：" + strconv.Itoa(resp.StatusCode))
	} else {
		return result, nil
	}
}

func Get(url string, param map[string]string, header map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// header头
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	// 参数
	if len(param) > 0 {
		query := req.URL.Query()
		for k, v := range param {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	return httpReqResp(req, url, param)
}

// Post 请求参数格式：json
func Post(url string, param []byte, header map[string]string) ([]byte, error) {
	body := bytes.NewBuffer(param)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	// header头
	req.Header.Set("Content-Type", "application/json")
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	return httpReqResp(req, url, param)
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(n int) string {
	b := make([]byte, n)
	length := int64(len(letterBytes))
	for i := range b {

		b[i] = letterBytes[rand.Int63()%length]
	}
	return string(b)
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	if size >= config.GB {
		return fmt.Sprintf("%.2f Gb", float64(size)/float64(config.GB))
	} else if size >= config.MB {
		return fmt.Sprintf("%.2f Mb", float64(size)/float64(config.MB))
	} else if size >= config.KB {
		return fmt.Sprintf("%.2f Kb", float64(size)/float64(config.KB))
	} else {
		return fmt.Sprintf("%d B", size)
	}
}

// SearchJsonByKeysRecursive 在json数据中递归查找指定键名相同的所有数据
func SearchJsonByKeysRecursive(object interface{}, key []string, handler func(object map[string]interface{}, key string)) {
	switch object.(type) {
	case []interface{}:
		object := object.([]interface{})
		for _, sub := range object {
			SearchJsonByKeysRecursive(sub, key, handler)
		}
	case map[string]interface{}:
		object := object.(map[string]interface{})
		for k, sub := range object {
			if SliceHas(key, k) {
				handler(object, k)
			} else {
				SearchJsonByKeysRecursive(sub, key, handler)
			}
		}
	}
}

// FormatStandardTime 格式化标准时间字符串
func FormatStandardTime(t string) string {
	if len(t) < 19 {
		return t
	}
	return strings.Replace(t, "T", " ", 1)[:19]
}

// SliceHas 值在切片中是否存在
func SliceHas[T int | string](list []T, value T) bool {
	if list == nil {
		return false
	}
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

// SliceUnique 切片去重
func SliceUnique[T int | string](array []T) []T {
	m := make(map[T]struct{})
	var ok bool
	var result []T
	for _, v := range array {
		if _, ok = m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

// SliceRemoveValue 按值删除切片元素
func SliceRemoveValue[T int | float64 | string](array []T, val T) []T {
	if len(array) == 0 {
		return array
	}
	var result []T
	for _, v := range array {
		if v != val {
			result = append(result, v)
		}
	}
	return result
}

// SliceAvg 求平均值
func SliceAvg[T int | float64](array []T) T {
	if len(array) == 0 {
		return T(0)
	}
	var sum T
	for _, v := range array {
		sum += v
	}
	return sum / T(len(array))
}

// SliceIntToString 数字切片转字符串切片
func SliceIntToString(sliceInt []int) []string {
	if len(sliceInt) == 0 {
		return nil
	}
	var sliceString []string
	for _, i := range sliceInt {
		sliceString = append(sliceString, strconv.Itoa(i))
	}
	return sliceString
}

// MapIntersect 求map[T]bool的交集
func MapIntersect[T int | string](a map[T]bool, b map[T]bool) map[T]bool {
	res := make(map[T]bool)
	for v := range a {
		if _, ok := b[v]; ok {
			res[v] = true
		}
	}
	return res
}
