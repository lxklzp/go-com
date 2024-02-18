package controller

import (
	"github.com/gin-gonic/gin"
	"go-com/global"
	"go-com/internal/model"
)

type rule struct {
}

func InitController() {}

func init() {
	global.AddRouterApi(rule{})
}

func (ctl rule) ActionAddStorage(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	if err = c.ShouldBindJSON(&m); err != nil {
		return global.RespData(400, err.Error(), nil)
	}

	return global.RespData(200, "ActionAddStorage", m)
}

func (ctl rule) ActionUpdStorage(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	if err = c.ShouldBindJSON(&m); err != nil {
		return global.RespData(400, err.Error(), nil)
	}

	return global.RespData(200, "ActionUpdStorage", nil)
}

func (ctl rule) ActionDelStorage(c *gin.Context) interface{} {
	var err error
	var param model.PrimaryId
	if err = c.ShouldBindJSON(&param); err != nil {
		return global.RespData(400, err.Error(), nil)
	}
	return global.RespData(200, "ActionDelStorage", nil)
}
