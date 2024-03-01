package model

import (
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go-com/config"
	"go-com/core/mod"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type RuleStorageCondition struct {
	IncludeAlarmKeyword string           `gorm:"column:include_alarm_keyword;comment:包含告警关键字" json:"include_alarm_keyword"`  // 包含告警关键字
	ExcludeAlarmKeyword string           `gorm:"column:exclude_alarm_keyword;comment:不包含告警关键字" json:"exclude_alarm_keyword"` // 不包含告警关键字
	RuleBeginTime       config.Timestamp `gorm:"column:rule_begin_time;comment:规则开始时间" json:"rule_begin_time"`               // 规则开始时间
	RuleEndTime         config.Timestamp `gorm:"column:rule_end_time;comment:规则结束时间" json:"rule_end_time"`                   // 规则结束时间
}

type RuleStorageResultPart struct {
	WayName    string `gorm:"-"`
	Way        string `gorm:"column:way;comment:,分隔 1 入库不处理 2 正常工单 3 短信单 4 仅智能网管发短信 5 隐患单" json:"way"` // ,分隔 1 入库不处理 2 正常工单 3 短信单 4 仅智能网管发短信 5 隐患单
	OrderDelay int32  `gorm:"column:order_delay;comment:派单延迟时间（分钟）" json:"order_delay"`                // 派单延迟时间（分钟）
}

type RuleStorageResult struct {
	RuleStorageResultPart

	ID           int64 `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Priority     int32 `gorm:"column:priority;comment:优先级" json:"priority"` // 优先级
	ConditionNum int8  `gorm:"-"`
}

type RuleStorageMatch struct {
	RuleUnitExt
	RuleStorageCondition
	RuleStorageResult
}

type RuleStorage struct {
	RuleUnit
	RuleStorageMatch
	UserID     int32            `gorm:"column:user_id" json:"user_id"`
	UserName   string           `gorm:"-" json:"user_name"`
	CreateTime config.Timestamp `gorm:"column:create_time;comment:创建时间" json:"create_time"` // 创建时间
	UpdateTime config.Timestamp `gorm:"column:update_time;comment:更新时间" json:"update_time"`
}

type RuleStorageHistory RuleStorage

type RuleStorageMatches []RuleStorageMatch

func (s RuleStorageMatches) Len() int {
	return len(s)
}

func (s RuleStorageMatches) Less(i, j int) bool {
	return s[i].Priority*1000+int32(s[i].ConditionNum) > s[j].Priority*1000+int32(s[j].ConditionNum)
}

func (s RuleStorageMatches) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

type ExtRuleStorage struct {
	RuleStorage
	mod.Base
}

func (m *RuleStorage) FormatList(list []RuleStorage, c *gin.Context) {
	var userIds []int32
	for k := range list {
		list[k].WayName = strings.TrimRight(list[k].WayName, ",")
		userIds = append(userIds, list[k].UserID)
	}
}

func (m *RuleStorage) queryList(param ExtRuleStorage) *gorm.DB {
	query := config.Pg.Table("rule_storage ll").Where("ll.id>0")
	mod.FilterWhere(query, "ll.id=?", param.ID)
	mod.FilterWhere(query, "ll.specialty=?", param.Specialty)
	mod.FilterWhere(query, "ll.net_manage=?", param.NetManage)
	mod.FilterWhere(query, "ll.city=?", param.City)
	mod.FilterWhere(query, "ll.area=?", param.Area)
	mod.FilterWhere(query, "ll.alarm_level=?", param.AlarmLevel)
	mod.FilterWhere(query, "ll.device_name=?", param.DeviceName)
	mod.FilterWhere(query, "ll.device_type=?", param.DeviceType)
	mod.FilterWhere(query, "ll.alarm_sub_type=?", param.AlarmSubType)
	mod.FilterWhere(query, "ll.include_alarm_keyword=?", param.IncludeAlarmKeyword)
	mod.FilterWhere(query, "ll.exclude_alarm_keyword=?", param.ExcludeAlarmKeyword)
	mod.FilterWhere(query, "ll.way=?", param.Way)
	mod.FilterWhere(query, "ll.order_delay=?", param.OrderDelay)
	mod.FilterWhere(query, "ll.rule_begin_time=?", param.RuleBeginTime)
	mod.FilterWhere(query, "ll.rule_end_time=?", param.RuleEndTime)
	mod.FilterWhere(query, "ll.priority=?", param.Priority)

	return query
}

func (m *RuleStorage) List(param ExtRuleStorage, isCount bool, c *gin.Context) map[string]interface{} {
	query := m.queryList(param)

	var count int64
	if isCount {
		query.Count(&count)
	}

	var list []RuleStorage
	param.Base.Validate()
	query.Select("ll.*").
		Order("ll.id desc").Limit(param.PageSize).Offset((param.Page - 1) * param.PageSize).Find(&list)
	m.FormatList(list, c)

	return map[string]interface{}{"list": list, "count": count}
}

// ExcelExport 使用时，有3处地方需要修改
func (m *RuleStorage) ExcelExport(param ExtRuleStorage, c *gin.Context) (string, error) {
	// 1 名称和标题
	return mod.ExcelExport("入库规则", []interface{}{"规则ID", "专业", "网管", "地市区域", "区县（子区域）", "告警级别", "设备名称", "设备类型", "告警子类型", "包含告警关键字", "不包含告警关键字", "规则开始时间", "规则结束时间", "处理方式", "派单延迟时间（分钟）", "优先级"},
		func(page int, pageSize int, isCount bool) mod.ExcelReadTable {
			param.Page = page
			param.PageSize = pageSize
			result := m.List(param, isCount, c)
			return mod.ExcelReadTable{result["count"].(int64), result["list"]}
		}, func(stream *excelize.StreamWriter, table mod.ExcelReadTable, rowNext *int) {
			var row []interface{}
			for _, item := range table.List.([]RuleStorage) { // 2 model类型
				// 3 表格字段
				row = []interface{}{item.ID, item.Specialty, item.NetManage, item.City, item.Area, item.AlarmLevelName, item.DeviceName, item.DeviceType, item.AlarmSubType, item.IncludeAlarmKeyword, item.ExcludeAlarmKeyword, item.RuleBeginTime.String(), item.RuleEndTime.String(), item.WayName, item.OrderDelay, item.Priority}
				stream.SetRow("A"+strconv.Itoa(*rowNext), row)
				*rowNext++
			}
		})
}
