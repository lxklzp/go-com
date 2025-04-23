package ctlResource

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/mod"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/model"
	"time"
)

/* 增删改查示例 */

type robotGroup struct {
}

func init() {
	config.AddRouterApi(robotGroup{}, &config.RouterWebApiList)
}

type robotGroupReqData struct {
	model.RobotGroup
	mod.Base
}

func (ctl robotGroup) ActionAdd(c *gin.Context) interface{} {
	var err error
	var param robotGroupReqData
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, err.Error(), nil)
	}

	if param.Name == "" {
		return tool.RespData(400, "name有误。", nil)
	}
	if param.PlatformID == 0 {
		return tool.RespData(400, "platform_id有误。", nil)
	}

	var m model.RobotGroup
	app.Db.Select("id").Where("name=?", param.Name).Take(&m)
	if m.ID != 0 {
		return tool.RespData(400, "名称已存在。", nil)
	}

	param.CreateTime = config.Timestamp(time.Now())
	if err = app.Db.Create(&param.RobotGroup).Error; err != nil {
		return tool.RespData(500, err.Error(), nil)
	}
	return tool.RespData(200, "", map[string]interface{}{"id": param.ID})
}

func (ctl robotGroup) ActionUpd(c *gin.Context) interface{} {
	var err error
	var param robotGroupReqData
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, err.Error(), nil)
	}

	if param.ID == 1 {
		return tool.RespData(400, "默认分组。", nil)
	}
	if param.Name == "" {
		return tool.RespData(400, "name有误。", nil)
	}

	var m model.RobotGroup
	app.Db.Select("id").Where("name=? and id!=?", param.Name, param.ID).Take(&m)
	if m.ID != 0 {
		return tool.RespData(400, "名称已存在。", nil)
	}

	if err = app.Db.Select("name").Updates(&param.RobotGroup).Error; err != nil {
		return tool.RespData(500, err.Error(), nil)
	}
	return tool.RespData(200, "", nil)
}

func (ctl robotGroup) ActionDel(c *gin.Context) interface{} {
	var err error
	var param robotGroupReqData
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, err.Error(), nil)
	}
	if param.ID == 1 {
		return tool.RespData(400, "默认分组。", nil)
	}

	if err = app.Db.Delete(&param.RobotGroup).Error; err != nil {
		return tool.RespData(500, err.Error(), nil)
	}

	return tool.RespData(200, "", nil)
}

func (ctl robotGroup) ActionList(c *gin.Context) interface{} {
	var err error
	var param robotGroupReqData
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, err.Error(), nil)
	}

	var list []model.RobotGroup
	query := app.Db.Model(model.RobotGroup{})
	if param.Name != "" {
		query.Where("name like ?", "%"+param.Name+"%")
	}
	if param.PlatformID != 0 {
		query.Where("platform_id=?", param.PlatformID)
	}
	var count int64
	query.Count(&count)
	param.Base.Validate()
	query.Order("id desc").Limit(param.PageSize).Offset((param.Page - 1) * param.PageSize).Find(&list)

	return tool.RespData(200, "", map[string]interface{}{"count": count, "list": list})
}
