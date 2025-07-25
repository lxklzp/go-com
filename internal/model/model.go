package model

import (
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/system/protocol"
)

const TableNameRobotGroup = "robot_group"

// RobotGroup mapped from table <robot_group>
type RobotGroup struct {
	ID         int            `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Name       string         `gorm:"column:name;not null" json:"name"`
	UserID     int            `gorm:"column:user_id" json:"user_id"`
	PlatformID int            `gorm:"column:platform_id;not null" json:"platform_id"`
	CreateTime tool.Timestamp `gorm:"column:create_time" json:"create_time"`
}

// TableName RobotGroup's table name
func (m *RobotGroup) TableName() string {
	return protocol.GetTableNameFull(TableNameRobotGroup)
}

const TableNameCode = "code"

// Code mapped from table <code>
type Code struct {
	Type       string         `gorm:"column:type;primaryKey" json:"type"`
	Key        string         `gorm:"column:key;primaryKey" json:"key"`
	Value      string         `gorm:"column:value" json:"value"`
	Content    string         `gorm:"column:content" json:"content"`
	PKey       string         `gorm:"column:p_key;primaryKey" json:"p_key"`
	CreateTime tool.Timestamp `gorm:"column:create_time" json:"create_time"`
	Comment    string         `gorm:"column:comment" json:"comment"`
}

// TableName Code's table name
func (*Code) TableName() string {
	return protocol.GetTableNameFull(TableNameCode)
}

func CodeMap(ty string) map[string]string {
	var mList []Code
	app.Db.Select("key,value").Where("type=?", ty).Find(&mList)
	mMap := make(map[string]string)
	for _, m := range mList {
		mMap[m.Key] = m.Value
	}
	return mMap
}

func CodeMapFull(ty string) map[string]Code {
	var mList []Code
	app.Db.Select("key,value,content,p_key,comment").Where("type=?", ty).Find(&mList)
	mMap := make(map[string]Code)
	for _, m := range mList {
		mMap[m.Key] = m
	}
	return mMap
}

func CodeKey(ty string) []string {
	var mList []Code
	app.Db.Select("key").Where("type=?", ty).Find(&mList)
	var mKey []string
	for _, m := range mList {
		mKey = append(mKey, m.Key)
	}
	return mKey
}

const TableNameDigitalLifeIndex = "digital_life_index"

// DigitalLifeIndex mapped from table <digital_life_index>
type DigitalLifeIndex struct {
	ID             int            `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	IndicatorID    int            `gorm:"column:indicator_id" json:"indicatorId"`
	ProductID      int            `gorm:"column:product_id" json:"product_id"`
	ProvinceName   string         `gorm:"column:province_name" json:"provinceName"`
	ProvinceCode   string         `gorm:"column:province_code" json:"provinceCode"`
	IndicatorValue string         `gorm:"column:indicator_value" json:"indicatorValue"`
	TimeType       int            `gorm:"column:time_type" json:"timeType"`
	Time           string         `gorm:"column:time" json:"time"`
	CreateTime     tool.Timestamp `gorm:"column:create_time" json:"create_time"`

	IndicatorName string `gorm:"-" json:"indicator_name"`
	ProductName   string `gorm:"-" json:"product_name"`
	TimeName      string `gorm:"-" json:"time_name"`
}

// TableName DigitalLifeIndex's table name
func (*DigitalLifeIndex) TableName() string {
	return protocol.GetTableNameFull(TableNameDigitalLifeIndex)
}

const TableNameAuthConfig = "auth_config"

// AuthConfig mapped from table <auth_config>
type AuthConfig struct {
	ID         int            `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Name       string         `gorm:"column:name;not null" json:"name"`
	Value      string         `gorm:"column:value;not null" json:"value"`
	Comment    string         `gorm:"column:comment;not null" json:"comment"`
	Type       string         `gorm:"column:type;not null;comment:int string text" json:"type"` // int string text
	UpdateTime tool.Timestamp `gorm:"column:update_time;not null" json:"update_time"`
	UserID     int            `gorm:"column:user_id" json:"user_id"`
}

// TableName AuthConfig's table name
func (m *AuthConfig) TableName() string {
	return protocol.GetTableNameFull(TableNameAuthConfig)
}
