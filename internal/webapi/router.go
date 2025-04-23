package webapi

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/tool"
	"go-com/internal/system/resource"
	"go-com/internal/webapi/controller"
	"go-com/internal/webapi/controller/ctlResource"
	"net/http"
)

func bind(r *gin.Engine) {
	controller.InitController()
	ctlResource.InitController()
	for _, routerApi := range config.RouterWebApiList {
		routerApi := routerApi
		routerApi.Path = config.WebApiPrefix + routerApi.Path
		r.POST(routerApi.Path, func(c *gin.Context) {
			c.JSON(http.StatusOK, routerApi.Action(c))
		})
	}

	r.GET(config.WebApiPrefix+"version", func(c *gin.Context) {
		c.String(http.StatusOK, config.Version)
	})
	r.GET(config.WebApiPrefix+"config", func(c *gin.Context) {
		if !resource.Header.CheckManageToken(c) {
			c.JSON(http.StatusOK, tool.RespData(400, "权限不足。", nil))
			return
		}
		c.JSON(http.StatusOK, config.C.App)
	})
}
