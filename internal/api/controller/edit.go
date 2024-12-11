package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"go-com/config"
	"go-com/core/mod"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/model"
	"time"
)

/* 增删改接口示例 */

type alarmConfig struct{}

func init() {
	config.AddRouterApi(alarmConfig{}, &config.RouterApiList)
}

type AlarmConfigReqData struct {
	model.AlarmConfig
	mod.Base
	EditData    []string `json:"edit_data"`
	OrderNbr    string   `json:"order_nbr"`
	City        string   `json:"city"`
	ProductCode string   `json:"product_code"`
}

func (ctl alarmConfig) ActionAlarmConfigAdd(c *gin.Context) interface{} {
	var err error
	var param AlarmConfigReqData
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, tool.ErrorStr(err), nil)
	}

	m := param.AlarmConfig
	m.CreateTime = config.Timestamp(time.Now())

	if err = app.Mysql.Create(&m).Error; err != nil {
		switch err.(*mysql.MySQLError).Number {
		case 1062:
			return tool.RespData(400, "环节已存在", nil)
		default:
			return tool.RespData(500, err.Error(), nil)
		}
	}
	return tool.RespData(200, "", nil)
}

func (ctl alarmConfig) ActionAlarmConfigUpd(c *gin.Context) interface{} {
	var err error

	var param map[string]interface{}
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, tool.ErrorStr(err), nil)
	}

	if err = app.Mysql.Model(model.AlarmConfig{}).Where("id=?", param["id"]).Updates(param).Error; err != nil {
		switch err.(*mysql.MySQLError).Number {
		case 1062:
			return tool.RespData(400, "环节已存在", nil)
		default:
			return tool.RespData(500, err.Error(), nil)
		}
	}
	return tool.RespData(200, "", nil)
}

func (ctl alarmConfig) ActionAlarmConfigUpdStatus(c *gin.Context) interface{} {
	var err error

	var param AlarmConfigReqData
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, tool.ErrorStr(err), nil)
	}

	if err = app.Mysql.Model(model.AlarmConfig{}).Where("id=?", param.ID).Update("status", param.Status).Error; err != nil {
		return tool.RespData(500, err.Error(), nil)
	}
	return tool.RespData(200, "", nil)
}

func (ctl alarmConfig) ActionAlarmConfigDel(c *gin.Context) interface{} {
	var err error

	var param AlarmConfigReqData
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, tool.ErrorStr(err), nil)
	}

	app.Mysql.Delete(&param.AlarmConfig)
	return tool.RespData(200, "", nil)
}
