package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/global"
	"go-com/internal/model"
	webApi "go-com/internal/webapi"
)

var Rule rule

type rule struct {
}

func init() {

}

func (ctl rule) ActionAddStorage(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	if err = c.ShouldBindJSON(&m); err != nil {
		return global.RespData(400, err.Error(), nil)
	}

	return global.RespData(200, "", nil)
}

func (ctl rule) ActionUpdStorage(c *gin.Context) interface{} {
	var err error
	var m model.RuleStorage
	if err = c.ShouldBindJSON(&m); err != nil {
		return global.RespData(400, err.Error(), nil)
	}

	return global.RespData(200, "", nil)
}

func (ctl rule) ActionDelStorage(c *gin.Context) interface{} {
	var err error
	var param model.PrimaryId
	if err = c.ShouldBindJSON(&param); err != nil {
		return global.RespData(400, err.Error(), nil)
	}
	return global.RespData(200, "", nil)
}

func main() {
	config.Load()
	global.InitDefine()
	global.InitLog("test")

	webApi.Run()

	// 保持主协程
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
