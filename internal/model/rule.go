package model

type RuleUnit struct {
	Specialty     int8   `gorm:"column:specialty;comment:专业" json:"specialty"` // 专业
	SpecialtyName string `gorm:"-"`
	NetManage     string `gorm:"column:net_manage;comment:网管" json:"net_manage"` // 网管
}

type RuleUnitExt struct {
	City           string `gorm:"column:city;comment:地市区域 编码" json:"city"`            // 地市区域 编码
	Area           string `gorm:"column:area;comment:区县（子区域）" json:"area"`            // 区县（子区域）
	AlarmLevel     int8   `gorm:"column:alarm_level;comment:告警级别" json:"alarm_level"` // 告警级别
	AlarmLevelName string `gorm:"-"`
	DeviceName     string `gorm:"column:device_name;comment:设备名称" json:"device_name"`        // 设备名称
	DeviceType     string `gorm:"column:device_type;comment:设备类型" json:"device_type"`        // 设备类型
	AlarmSubType   string `gorm:"column:alarm_sub_type;comment:告警子类型" json:"alarm_sub_type"` // 告警子类型
}

type RuleUnique struct {
	RuleUnit
	RuleUnitExt
}
