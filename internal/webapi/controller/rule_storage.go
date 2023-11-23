package controller

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/global"
	"go-com/internal/model"
	"go-com/internal/webapi/rpc"
	"gorm.io/gorm"
	"time"
)

var RuleStorage ruleStorage

type ruleStorage struct {
}

func (ctl *ruleStorage) ActionAdd(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	if err = c.ShouldBindJSON(&m); err != nil {
		return global.RespData(400, err.Error(), nil)
	}

	if _, ok := config.CityIndex[m.City]; !ok && m.City != "" {
		return global.RespData(400, "City 参数错误", nil)
	}
	if m.OrderDelay < 2 {
		return global.RespData(400, "OrderDelay 需要大于等于2", nil)
	}

	m.UserID = rpc.Auth.CurrentUserId(c.Request)
	m.CreateTime = global.Timestamp(time.Now())
	m.UpdateTime = m.CreateTime

	if err = global.GormPg.Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(&m).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return global.RespData(500, err.Error(), nil)
	}

	return global.RespData(200, "", nil)
}

func (ctl *ruleStorage) ActionUpd(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	if err = c.ShouldBindJSON(&m); err != nil {
		return global.RespData(400, err.Error(), nil)
	}

	if _, ok := config.CityIndex[m.City]; !ok && m.City != "" {
		return global.RespData(400, "City 参数错误", nil)
	}

	m.UserID = rpc.Auth.CurrentUserId(c.Request)
	m.UpdateTime = global.Timestamp(time.Now())

	if err = global.GormPg.Transaction(func(tx *gorm.DB) error {
		if err = tx.Omit("create_time").Save(&m).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return global.RespData(500, err.Error(), nil)
	}

	return global.RespData(200, "", nil)
}

func (ctl *ruleStorage) ActionList(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	var param model.ExtRuleStorage
	if err = c.ShouldBindJSON(&param); err != nil {
		return global.RespData(400, err.Error(), nil)
	}
	data := m.List(param, true, c)

	data["time_min"] = global.DefaultTimeMin
	data["time_max"] = global.DefaultTimeMax
	return global.RespData(200, "", data)
}

func (ctl *ruleStorage) ActionExport(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	var param model.ExtRuleStorage
	if err = c.ShouldBindJSON(&param); err != nil {
		return global.RespData(400, err.Error(), nil)
	}
	path, err := m.ExcelExport(param, c)
	if err != err {
		return global.RespData(500, err.Error(), nil)
	}

	return global.RespData(200, "", map[string]interface{}{"path": path})
}

func (ctl *ruleStorage) ActionDel(c *gin.Context) interface{} {
	var err error
	var param model.PrimaryId
	if err = c.ShouldBindJSON(&param); err != nil {
		return global.RespData(400, err.Error(), nil)
	}

	if err = global.GormPg.Transaction(func(tx *gorm.DB) error {
		if err = tx.Model(model.RuleStorage{}).Delete("id=?", param.ID).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		return global.RespData(500, err.Error(), nil)
	}

	return global.RespData(200, "", nil)
}
