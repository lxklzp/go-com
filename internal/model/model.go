package model

import (
	"go-com/config"
	"go-com/internal/app"
)

const TableNameCode = `"zhan_xin"."code"`

// Code mapped from table <code>
type Code struct {
	Type       string           `gorm:"column:type;primaryKey" json:"type"`
	Key        string           `gorm:"column:key;primaryKey" json:"key"`
	Value      string           `gorm:"column:value" json:"value"`
	Content    string           `gorm:"column:content" json:"content"`
	PKey       string           `gorm:"column:p_key;primaryKey" json:"p_key"`
	CreateTime config.Timestamp `gorm:"column:create_time" json:"create_time"`
	Comment    string           `gorm:"column:comment" json:"comment"`
}

// TableName Code's table name
func (*Code) TableName() string {
	return TableNameCode
}

func CodeMap(ty string) map[string]string {
	var mList []Code
	app.Pg.Select("key,value").Where("type=?", ty).Find(&mList)
	mMap := make(map[string]string)
	for _, m := range mList {
		mMap[m.Key] = m.Value
	}
	return mMap
}

func CodeMapFull(ty string) map[string]Code {
	var mList []Code
	app.Pg.Select("key,value,content,p_key,comment").Where("type=?", ty).Find(&mList)
	mMap := make(map[string]Code)
	for _, m := range mList {
		mMap[m.Key] = m
	}
	return mMap
}

func CodeKey(ty string) []string {
	var mList []Code
	app.Pg.Select("key").Where("type=?", ty).Find(&mList)
	var mKey []string
	for _, m := range mList {
		mKey = append(mKey, m.Key)
	}
	return mKey
}

const TableNameDownload = `"zhan_xin"."download"`

// Download mapped from table <download>
type Download struct {
	ID         int              `gorm:"column:id;primaryKey" json:"id"`
	Name       string           `gorm:"column:name" json:"name"`
	Path       string           `gorm:"column:path" json:"path"`
	UserID     int              `gorm:"column:user_id" json:"user_id"`
	BeginTime  config.Timestamp `gorm:"column:begin_time" json:"begin_time"`
	EndTime    config.Timestamp `gorm:"column:end_time" json:"end_time"`
	CreateTime config.Timestamp `gorm:"column:create_time" json:"create_time"`
	Status     int              `gorm:"column:status" json:"status"`
}

// TableName Download's table name
func (*Download) TableName() string {
	return TableNameDownload
}
