package tool

import (
	"go-com/config"
	"go-com/core/logr"
	"gorm.io/gorm"
	"time"
)

// 节假日数据来源 https://www.gov.cn/gongbao/2023/issue_10806/202311/content_6913823.html

/*
CREATE TABLE `holiday` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `year` varchar(4) NOT NULL COMMENT '年份',
  `day` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin NOT NULL COMMENT '日期',
  `festival_type` tinyint(4) NOT NULL COMMENT '0 不用该字段 1 节假日 2 工作日（调休）',
  `holiday_type` tinyint(4) NOT NULL COMMENT '1 节假调休 2 全年休息日（节假日+周末）',
  `create_time` datetime DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `holiday_unique` (`year`,`day`,`holiday_type`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4;
*/

type Holiday struct {
	ID           int       `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Year         string    `gorm:"column:year;comment:年份" json:"year"`                                        // 年份
	Day          string    `gorm:"column:day;comment:日期" json:"day"`                                          // 日期
	FestivalType int       `gorm:"column:festival_type;comment:0 不用该字段 1 节假日 2 工作日（调休）" json:"festival_type"` // 0 不用该字段 1 节假日 2 工作日（调休）
	HolidayType  int       `gorm:"column:holiday_type;comment:1 节假调休 2 全年休息日（节假日+周末）" json:"holiday_type"`    // 1 节假调休 2 全年休息日（节假日+周末）
	CreateTime   Timestamp `gorm:"column:create_time" json:"create_time"`
}

// 获取一年中，法定节假日和周末日期
func generateHoliday(festival []string, workday []string) []string {
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

// 刷新数据库中节假日记录
func refreshHoliday(db *gorm.DB, year string, festival []string, workday []string) {
	// 先删除
	db.Where("year=?", year).Delete(Holiday{})

	now := time.Now()

	// 录入节假调休
	var list []Holiday
	for _, d := range festival {
		m := Holiday{
			Year:         year,
			Day:          d,
			FestivalType: 1,
			HolidayType:  1,
			CreateTime:   Timestamp(now),
		}
		list = append(list, m)
	}
	for _, d := range workday {
		m := Holiday{
			Year:         year,
			Day:          d,
			FestivalType: 2,
			HolidayType:  1,
			CreateTime:   Timestamp(now),
		}
		list = append(list, m)
	}
	db.Create(&list)

	// 录入全年休息日（节假日+周末）
	holiday := generateHoliday(festival, workday)
	list = list[:0]
	for _, d := range holiday {
		m := Holiday{
			Year:         year,
			Day:          d,
			FestivalType: 0,
			HolidayType:  2,
			CreateTime:   Timestamp(now),
		}
		list = append(list, m)
	}
	db.Create(&list)
}

func RefreshHoliday2025(db *gorm.DB) {
	refreshHoliday(db, "2025", []string{
		"2025-01-01",
		"2025-01-28",
		"2025-01-29",
		"2025-01-30",
		"2025-01-31",
		"2025-02-01",
		"2025-02-02",
		"2025-02-03",
		"2025-02-04",
		"2025-04-04",
		"2025-04-05",
		"2025-04-06",
		"2025-05-01",
		"2025-05-02",
		"2025-05-03",
		"2025-05-04",
		"2025-05-05",
		"2025-05-31",
		"2025-06-01",
		"2025-06-02",
		"2025-10-01",
		"2025-10-02",
		"2025-10-03",
		"2025-10-04",
		"2025-10-05",
		"2025-10-06",
		"2025-10-07",
		"2025-10-08",
	}, []string{
		"2025-01-26",
		"2025-02-08",
		"2025-04-27",
		"2025-09-28",
		"2025-10-11",
	})
}
