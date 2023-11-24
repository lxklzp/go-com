package api

import (
	"github.com/gin-gonic/gin"
	"go-com/global"
	"go-com/internal/api/controller"
	"net/http"
)

func bind(r *gin.Engine) {
	controller.InitController()
	for _, routerApi := range global.RouterApiList {
		routerApi := routerApi
		r.POST(routerApi.Path, func(c *gin.Context) {
			c.JSON(http.StatusOK, routerApi.Action(c))
		})
	}
}
