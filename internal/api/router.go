package api

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/tool"
	"go-com/internal/api/controller"
	"go-com/internal/system/resource"
	"net/http"
)

func bind(r *gin.Engine) {
	controller.InitController()
	for _, routerApi := range config.RouterAppApiList {
		routerApi := routerApi
		routerApi.Path = config.AppApiPrefix + routerApi.Path
		r.POST(routerApi.Path, func(c *gin.Context) {
			c.JSON(http.StatusOK, routerApi.Action(c))
		})
	}

	r.GET(config.AppApiPrefix+"version", func(c *gin.Context) {
		c.String(http.StatusOK, config.Version)
	})
	r.GET(config.AppApiPrefix+"config", func(c *gin.Context) {
		if !resource.Header.CheckManageToken(c) {
			c.JSON(http.StatusOK, tool.RespData(400, "权限不足。", nil))
			return
		}
		c.JSON(http.StatusOK, config.C.App)
	})
}
