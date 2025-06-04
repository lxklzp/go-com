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

// AlarmConfig 告警等级配置
type AlarmConfig struct {
	ID                 int            `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Type               string         `gorm:"column:type;not null;default:1;comment:1 集团 2 省" json:"type"` // 1 集团 2 省
	Product            string         `gorm:"column:product;not null;comment:产品code" json:"product"`       // 产品code
	OperateType        string         `gorm:"column:operate_type;not null;comment:操作类型/操作动作" json:"operate_type"`
	SegmentID          string         `gorm:"column:segment_id;not null;comment:环节ID" json:"segment_id"`             // 环节ID
	SegmentName        string         `gorm:"column:segment_name;comment:环节名称" json:"segment_name"`                  // 环节名称
	DataFrom           string         `gorm:"column:data_from;comment:数据来源" json:"data_from"`                        // 数据来源
	TimeoutPre         int            `gorm:"column:timeout_pre;comment:即将超时 分钟" json:"timeout_pre"`                 // 即将超时 分钟
	Timeout            int            `gorm:"column:timeout;comment:超时 分钟" json:"timeout"`                           // 超时 分钟
	IsAuto             int            `gorm:"column:is_auto;comment:超时自动派发工单：1 是 0 否" json:"is_auto"`                // 超时自动派发工单：1 是 0 否
	HandlerType        int            `gorm:"column:handler_type;comment:1 环节自动 2 手动处理 3 半自动处理" json:"handler_type"` // 1 环节自动 2 手动处理 3 半自动处理
	HandlerSys         string         `gorm:"column:handler_sys;comment:派往系统" json:"handler_sys"`                    // 派往系统
	Status             int            `gorm:"column:status;comment:1 启用 0 停用" json:"status"`                         // 1 启用 0 停用
	UserID             int            `gorm:"column:user_id" json:"user_id"`
	CreateTime         tool.Timestamp `gorm:"column:create_time" json:"create_time"`
	ChangeType         string         `gorm:"column:change_type;comment:变更类型" json:"change_type"`                               // 变更类型
	ProvincialContacts string         `gorm:"column:provincial_contacts;comment:卡单处理联系人" json:"provincial_contacts"`            // 卡单处理联系人
	ProvincialNum      string         `gorm:"column:provincial_num;comment:卡单处理联系人手机号" json:"provincial_num"`                   // 卡单处理联系人手机号
	ProvincialSwitch   int            `gorm:"column:provincial_switch;not null;comment:短信通知 1开启 0 关闭" json:"provincial_switch"` // 短信通知 1开启 0 关闭

	ProductName     string `gorm:"-" json:"product_name"`
	OperateTypeName string `gorm:"-" json:"operate_type_name"`
}
