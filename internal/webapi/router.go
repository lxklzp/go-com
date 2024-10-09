package webapi

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/internal/webapi/controller"
	"net/http"
)

func bind(r *gin.Engine) {
	controller.InitController()
	for _, routerApi := range config.RouterApiList {
		routerApi := routerApi
		r.POST(routerApi.Path, func(c *gin.Context) {
			c.JSON(http.StatusOK, routerApi.Action(c))
		})
	}

	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, config.Version)
	})

	// excel导出示例
	r.POST("/device-config-audit/export", func(c *gin.Context) {
		controller.Test.ActExport(c)
	})
}
