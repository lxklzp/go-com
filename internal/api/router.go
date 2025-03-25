package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/internal/api/controller"
	"net/http"
)

func bind(r *gin.Engine) {
	controller.InitController()
	for _, routerApi := range config.RouterApiList {
		routerApi := routerApi
		r.POST(routerApi.Path, func(c *gin.Context) {
			respData := routerApi.Action(c)
			if respData != nil {
				c.JSON(http.StatusOK, respData)
			}
		})
	}

	r.GET("/version", func(c *gin.Context) {
		c.String(http.StatusOK, config.Version)
	})
	r.Any("/nifi-api/*api", controller.Nifi.ProxyApi)
	r.GET("/nifi-phoenix-home/:header", func(c *gin.Context) {
		fmt.Println(c.Param("header"))
	})
}
