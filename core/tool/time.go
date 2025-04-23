package tool

import (
	"github.com/pkg/errors"
	"go-com/config"
	"strings"
	"time"
)

// FormatFromTimeStandard 格式化标准时间字符串
func FormatFromTimeStandard(t string) string {
	if len(t) < 19 {
		return t
	}
	return strings.Replace(t, "T", " ", 1)[:19]
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

// TimeFromToSeconds 获取两个时间段之间的时间，单位秒
func TimeFromToSeconds(timeFromStr string, timeToStr string) int {
	if timeFromStr == "" || timeToStr == "" {
		return 0
	} else {
		var timeFrom, timeTo time.Time
		timeFrom, _ = time.ParseInLocation(config.DateTimeFormatter, timeFromStr, time.Local)
		timeTo, _ = time.ParseInLocation(config.DateTimeFormatter, timeToStr, time.Local)
		return int(timeTo.Sub(timeFrom).Seconds())
	}
}
