package model

import "go-com/config"

const TableNameDigitalLifeIndex = `"zhan_xin"."digital_life_index"`

// DigitalLifeIndex mapped from table <digital_life_index>
type DigitalLifeIndex struct {
	ID             int              `gorm:"column:id;primaryKey" json:"id"`
	IndicatorID    int              `gorm:"column:indicator_id" json:"indicatorId"`
	ProductID      int              `gorm:"column:product_id" json:"product_id"`
	ProvinceName   string           `gorm:"column:province_name" json:"provinceName"`
	ProvinceCode   string           `gorm:"column:province_code" json:"provinceCode"`
	IndicatorValue string           `gorm:"column:indicator_value" json:"indicatorValue"`
	TimeType       int              `gorm:"column:time_type" json:"timeType"`
	Time           string           `gorm:"column:time" json:"time"`
	CreateTime     config.Timestamp `gorm:"column:create_time" json:"create_time"`

	IndicatorName string `gorm:"-" json:"indicator_name"`
	ProductName   string `gorm:"-" json:"product_name"`
	TimeName      string `gorm:"-" json:"time_name"`
}

// TableName DigitalLifeIndex's table name
func (*DigitalLifeIndex) TableName() string {
	return TableNameDigitalLifeIndex
}

// AlarmConfig 告警等级配置
type AlarmConfig struct {
	ID                 int              `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Type               string           `gorm:"column:type;not null;default:1;comment:1 集团 2 省" json:"type"` // 1 集团 2 省
	Product            string           `gorm:"column:product;not null;comment:产品code" json:"product"`       // 产品code
	OperateType        string           `gorm:"column:operate_type;not null;comment:操作类型/操作动作" json:"operate_type"`
	SegmentID          string           `gorm:"column:segment_id;not null;comment:环节ID" json:"segment_id"`             // 环节ID
	SegmentName        string           `gorm:"column:segment_name;comment:环节名称" json:"segment_name"`                  // 环节名称
	DataFrom           string           `gorm:"column:data_from;comment:数据来源" json:"data_from"`                        // 数据来源
	TimeoutPre         int              `gorm:"column:timeout_pre;comment:即将超时 分钟" json:"timeout_pre"`                 // 即将超时 分钟
	Timeout            int              `gorm:"column:timeout;comment:超时 分钟" json:"timeout"`                           // 超时 分钟
	IsAuto             int              `gorm:"column:is_auto;comment:超时自动派发工单：1 是 0 否" json:"is_auto"`                // 超时自动派发工单：1 是 0 否
	HandlerType        int              `gorm:"column:handler_type;comment:1 环节自动 2 手动处理 3 半自动处理" json:"handler_type"` // 1 环节自动 2 手动处理 3 半自动处理
	HandlerSys         string           `gorm:"column:handler_sys;comment:派往系统" json:"handler_sys"`                    // 派往系统
	Status             int              `gorm:"column:status;comment:1 启用 0 停用" json:"status"`                         // 1 启用 0 停用
	UserID             int              `gorm:"column:user_id" json:"user_id"`
	CreateTime         config.Timestamp `gorm:"column:create_time" json:"create_time"`
	ChangeType         string           `gorm:"column:change_type;comment:变更类型" json:"change_type"`                               // 变更类型
	ProvincialContacts string           `gorm:"column:provincial_contacts;comment:卡单处理联系人" json:"provincial_contacts"`            // 卡单处理联系人
	ProvincialNum      string           `gorm:"column:provincial_num;comment:卡单处理联系人手机号" json:"provincial_num"`                   // 卡单处理联系人手机号
	ProvincialSwitch   int              `gorm:"column:provincial_switch;not null;comment:短信通知 1开启 0 关闭" json:"provincial_switch"` // 短信通知 1开启 0 关闭
	ProductName        string           `gorm:"-" json:"product_name"`
	OperateTypeName    string           `gorm:"-" json:"operate_type_name"`
}
