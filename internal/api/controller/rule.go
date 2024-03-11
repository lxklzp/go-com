package controller

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/tool"
	"go-com/internal/model"
)

type rule struct {
}

func InitController() {}

func init() {
	config.AddRouterApi(rule{}, &config.RouterApiList)
}

func (ctl rule) ActionAddStorage(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	if err = c.ShouldBindJSON(&m); err != nil {
		return tool.RespData(400, err.Error(), nil)
	}

	return tool.RespData(200, "ActionAddStorage", m)
}

func (ctl rule) ActionUpdStorage(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	if err = c.ShouldBindJSON(&m); err != nil {
		return tool.RespData(400, err.Error(), nil)
	}

	return tool.RespData(200, "ActionUpdStorage", nil)
}
