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
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode"
)

func init() {
	regFloat, _ = regexp.Compile(`^[\-\d.]+`)
	regInt, _ = regexp.Compile(`^[\-\d]+`)
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
	if field == "" {
		return ""
	}
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

func SepNameToCamel(field string, isUcFirst bool) string {
	if field == "" {
		return ""
	}
	name := strings.ReplaceAll(cases.Title(language.English).String(strings.ReplaceAll(strings.ToLower(field), "_", " ")), " ", "")
	if isUcFirst {
		return name
	}
	return strings.ToLower(name[:1]) + name[1:]
}

func httpReqResp(req *http.Request, url string, param interface{}) ([]byte, error) {
	// 请求
	client := &http.Client{}
	client.Timeout = time.Minute
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

// SearchJsonOnceByKey 在json数据中查找一次指定键名的值
func SearchJsonOnceByKey(object interface{}, key string) interface{} {
	switch object.(type) {
	case []interface{}:
		object := object.([]interface{})
		for _, sub := range object {
			return SearchJsonOnceByKey(sub, key)
		}
	case map[string]interface{}:
		object := object.(map[string]interface{})
		for k, sub := range object {
			if k == key {
				return sub
			} else {
				if value := SearchJsonOnceByKey(sub, key); value != config.Sep {
					return value
				}
			}
		}
	}
	return config.Sep
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
	if len(array) == 0 {
		return nil
	}

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
		return nil
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

var regFloat, regInt *regexp.Regexp

// InterfaceToFloat interface转float
func InterfaceToFloat(v interface{}) (float64, error) {
	if v == nil {
		return 0, errors.New("不支持的类型")
	}
	switch v.(type) {
	case string:
		return strconv.ParseFloat(regFloat.FindString(v.(string)), 10)
	case float64:
		return v.(float64), nil
	case int:
		return float64(v.(int)), nil
	case int8:
		return float64(v.(int8)), nil
	case int16:
		return float64(v.(int16)), nil
	case int32:
		return float64(v.(int32)), nil
	case int64:
		return float64(v.(int64)), nil
	case uint:
		return float64(v.(uint)), nil
	case uint8:
		return float64(v.(uint8)), nil
	case uint16:
		return float64(v.(uint16)), nil
	case uint32:
		return float64(v.(uint32)), nil
	case uint64:
		return float64(v.(uint64)), nil
	default:
		return 0, errors.New("不支持的类型")
	}
}

// InterfaceToInt interface转int
func InterfaceToInt(v interface{}) (int, error) {
	if v == nil {
		return 0, errors.New("不支持的类型")
	}
	switch v.(type) {
	case string:
		return strconv.Atoi(regInt.FindString(v.(string)))
	case float64:
		return int(v.(float64)), nil
	case int:
		return v.(int), nil
	case int8:
		return int(v.(int8)), nil
	case int16:
		return int(v.(int16)), nil
	case int32:
		return int(v.(int32)), nil
	case int64:
		return int(v.(int64)), nil
	case uint:
		return int(v.(uint)), nil
	case uint8:
		return int(v.(uint8)), nil
	case uint16:
		return int(v.(uint16)), nil
	case uint32:
		return int(v.(uint32)), nil
	case uint64:
		return int(v.(uint64)), nil
	default:
		return 0, errors.New("不支持的类型")
	}
}

// 根据float排序，升序
type sortFloat struct {
	value float64
	index int
}
type SortFloatList []sortFloat

func (s SortFloatList) Len() int           { return len(s) }
func (s SortFloatList) Less(i, j int) bool { return s[i].value < s[j].value }
func (s SortFloatList) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// MapSortFloat 对map数组进行float排序，sortType：1 升序 2 降序
func MapSortFloat(mList []map[string]interface{}, sortField string, sortType int) []map[string]interface{} {
	// 参数验证
	if len(mList) == 0 || sortField == "" || (sortType != 1 && sortType != 2) {
		return nil
	}

	// 准备排序数据
	var ok bool
	var err error
	var value, valueDefault float64
	if sortType == 1 {
		valueDefault = config.MaxFloat
	} else {
		valueDefault = config.MinFloat
	}
	// 构建排序结构体：SortFloatList
	var sortFloatList SortFloatList
	sortFloatList = make([]sortFloat, 0, len(mList))
	for k, m := range mList {
		if _, ok = m[sortField]; !ok {
			// 处理排序字段不存在的情况
			value = valueDefault
		} else if value, err = InterfaceToFloat(m[sortField]); err != nil {
			// 处理排序字段不能转换成float64的情况
			value = valueDefault
		}
		sortFloatList = append(sortFloatList, sortFloat{
			value: value,
			index: k,
		})
	}

	// SortFloatList排序
	if sortType == 1 {
		sort.Sort(sortFloatList)
	} else {
		sort.Sort(sort.Reverse(sortFloatList))
	}

	// 根据已排序的SortFloatList，生成结果数据
	result := make([]map[string]interface{}, 0, len(mList))
	for _, si := range sortFloatList {
		result = append(result, mList[si.index])
	}

	return result
}

// 根据string排序，升序
type sortString struct {
	value string
	index int
}
type SortStringList []sortString

func (s SortStringList) Len() int { return len(s) }
func (s SortStringList) Less(i, j int) bool {
	if strings.Compare(s[i].value, s[j].value) < 0 {
		return true
	} else {
		return false
	}
}
func (s SortStringList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// MapSortString 对map数组进行float排序，sortType：1 升序 2 降序
func MapSortString(mList []map[string]interface{}, sortField string, sortType int) []map[string]interface{} {
	// 参数验证
	if len(mList) == 0 || sortField == "" || (sortType != 1 && sortType != 2) {
		return nil
	}

	// 准备排序数据
	var ok bool
	var value string
	// 构建排序结构体：SortStringList
	var sortStringList SortStringList
	sortStringList = make([]sortString, 0, len(mList))
	for k, m := range mList {
		if _, ok = m[sortField]; !ok {
			// 处理排序字段不存在的情况
			value = ""
		} else {
			value = InterfaceToString(m[sortField])
		}
		sortStringList = append(sortStringList, sortString{
			value: value,
			index: k,
		})
	}

	// SortStringList排序
	if sortType == 1 {
		sort.Sort(sortStringList)
	} else {
		sort.Sort(sort.Reverse(sortStringList))
	}

	// 根据已排序的SortFloatList，生成结果数据
	result := make([]map[string]interface{}, 0, len(mList))
	for _, si := range sortStringList {
		result = append(result, mList[si.index])
	}

	return result
}

// Holiday 获取一年中，法定节假日和周末日期
// 节假日数据来源 https://www.gov.cn/gongbao/2023/issue_10806/202311/content_6913823.html
func Holiday(festival []string, workday []string) []string {
	var holiday []string
	if len(festival) == 0 || len(workday) == 0 {
		logr.L.Error("节假日数据有误")
		return holiday
	}

	// 周末日期
	var weekend []string
	year := festival[0][:4]
	end, _ := time.ParseInLocation(config.DateTimeFormatter, year+"-12-31 00:00:01", time.Local)
	for step, _ := time.ParseInLocation(config.DateTimeFormatter, year+"-01-01 00:00:00", time.Local); step.Before(end); step = step.Add(time.Hour * 24) {
		weekday := int(step.Weekday())
		day := step.Format(config.DateFormatter)
		if (weekday == 6 || weekday == 0) && (!SliceHas(workday, day)) {
			weekend = append(weekend, step.Format(config.DateFormatter))
		}
	}

	holiday = append(holiday, festival...)
	holiday = append(holiday, weekend...)
	holiday = SliceUnique(holiday)
	return holiday
}

// GetTimeFromToPeriodExceptHoliday 获取两个时间段之间排除节假日的时间，单位秒
func GetTimeFromToPeriodExceptHoliday(timeFromStr string, timeToStr string, holiday []string) (int, error) {
	var err error
	var timeFrom, timeTo, dateFrom, dateTo, dateStep time.Time

	// 验证请求参数
	if timeFrom, err = time.ParseInLocation(config.DateTimeFormatter, timeFromStr, time.Local); err != nil {
		return 0, err
	}
	if timeTo, err = time.ParseInLocation(config.DateTimeFormatter, timeToStr, time.Local); err != nil {
		return 0, err
	}
	if timeFrom.After(timeTo) {
		return 0, errors.New("开始时间在结束时间之后")
	}

	// 开始、结束时间是同一天
	if timeFromStr[:10] == timeToStr[:10] {
		if SliceHas(holiday, timeFromStr[:10]) {
			return 0, nil
		} else {
			return int(timeTo.Sub(timeFrom) / time.Second), nil
		}
	}

	// 开始、结束时间不是同一天
	var period int
	dateFrom, _ = time.ParseInLocation(config.DateFormatter, timeFromStr[:10], time.Local)
	dateTo, _ = time.ParseInLocation(config.DateFormatter, timeToStr[:10], time.Local)

	// 开始当天
	if !SliceHas(holiday, timeFromStr[:10]) {
		period += int(dateFrom.AddDate(0, 0, 1).Sub(timeFrom) / time.Second)
	}
	// 中间日期
	dateStep = dateFrom.AddDate(0, 0, 1)
	for dateStep.Before(dateTo) {
		if !SliceHas(holiday, dateStep.Format(config.DateFormatter)) {
			period += 3600 * 24
		}
		dateStep = dateStep.AddDate(0, 0, 1)
	}
	// 结束当天
	if !SliceHas(holiday, timeToStr[:10]) {
		period += int(timeTo.Sub(dateTo) / time.Second)
	}
	return period, nil
}
