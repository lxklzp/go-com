package controller

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/gw"
	"go-com/core/mod"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/model"
	"strconv"
	"time"
)

type hehe struct {
}

func InitController() {}

func init() {
	config.AddRouterApi(hehe{}, &config.RouterApiWebList)
}

func (ctl hehe) ActionAdd(c *gin.Context) interface{} {
	var err error
	var m model.Hehe
	if err = c.ShouldBindJSON(&m); err != nil {
		return tool.RespData(400, tool.ErrorStr(err), nil)
	}

	var mName mod.Name
	app.Pg.Model(m).Where("name=?", m.Name).Take(&mName)
	if mName.Name != "" {
		return tool.RespData(400, "名称已存在", nil)
	}

	m.UserID, _ = strconv.Atoi(c.Request.Header.Get("user-id"))
	m.CreateTime = config.Timestamp(time.Now())
	if err = app.Pg.Create(&m).Error; err != nil {
		return tool.RespData(500, err.Error(), nil)
	}
	return tool.RespData(200, "", nil)
}

func (ctl hehe) ActionUpd(c *gin.Context) interface{} {
	var err error
	var m model.Hehe
	if err = c.ShouldBindJSON(&m); err != nil {
		return tool.RespData(400, tool.ErrorStr(err), nil)
	}

	var mName mod.Name
	app.Pg.Model(m).Where("name=? and id!=?", m.Name, m.ID).Take(&mName)
	if mName.Name != "" {
		return tool.RespData(400, "名称已存在", nil)
	}

	m.UserID, _ = strconv.Atoi(c.Request.Header.Get("user-id"))
	if err = app.Pg.Omit("create_time").Updates(&m).Error; err != nil {
		return tool.RespData(500, err.Error(), nil)
	}
	return tool.RespData(200, "", nil)
}

func (ctl hehe) ActionList(c *gin.Context) interface{} {
	var err error
	var param struct {
		Name string
		mod.Base
	}
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, err.Error(), nil)
	}

	var list []model.Hehe
	query := app.Pg.Model(model.Hehe{})
	if param.Name != "" {
		query.Where("name like ?", "%"+param.Name+"%")
	}
	var count int64
	query.Count(&count)
	param.Base.Validate()
	query.Order("id desc").Limit(param.PageSize).Offset((param.Page - 1) * param.PageSize).Find(&list)

	return tool.RespData(200, "", map[string]interface{}{"count": count, "list": list})
}

func (ctl hehe) ActionDel(c *gin.Context) interface{} {
	var err error
	var param mod.PrimaryId
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, err.Error(), nil)
	}

	if err = app.Pg.Where("id=?", param.ID).Delete(model.Hehe{}).Error; err != nil {
		return tool.RespData(500, err.Error(), nil)
	}

	return tool.RespData(200, "", nil)
}

func (ctl hehe) ActionUpload(c *gin.Context) interface{} {
	var result map[string]interface{}
	var err error
	if result, err = gw.Upload(c); err != nil {
		return tool.RespData(500, err.Error(), nil)
	}

	return tool.RespData(200, "", result)
}
