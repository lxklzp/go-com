package model

import (
	"go-com/config"
)

type Hehe struct {
	Name       string           `gorm:"column:name;not null" json:"name"`
	Age        int              `gorm:"column:age;not null" json:"age"`
	ID         int64            `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UserID     int              `gorm:"column:user_id" json:"user_id"`
	CreateTime config.Timestamp `gorm:"column:create_time" json:"create_time"`
}
