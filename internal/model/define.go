package model

import (
	"go-com/core/tool"
	"go-com/internal/system/protocol"
)

const TableNameDownload = "download"

// Download mapped from table <download>
type Download struct {
	ID         int            `gorm:"column:id;primaryKey" json:"id"`
	Name       string         `gorm:"column:name" json:"name"`
	Path       string         `gorm:"column:path" json:"path"`
	UserID     int            `gorm:"column:user_id" json:"user_id"`
	BeginTime  tool.Timestamp `gorm:"column:begin_time" json:"begin_time"`
	EndTime    tool.Timestamp `gorm:"column:end_time" json:"end_time"`
	CreateTime tool.Timestamp `gorm:"column:create_time" json:"create_time"`
	Status     int            `gorm:"column:status;comment:1 下载中 2 下载成功 3 下载失败" json:"status"`
}

// TableName Download's table name
func (*Download) TableName() string {
	return protocol.GetTableNameFull(TableNameDownload)
}
