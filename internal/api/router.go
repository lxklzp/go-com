package api

import (
	"github.com/gin-gonic/gin"
	"go-com/internal/api/controller"
	"net/http"
)

func bind(r *gin.Engine) {
	r.POST("rule/add-storage", func(c *gin.Context) {
		c.JSON(http.StatusOK, controller.Rule.ActionAddStorage(c))
	})
	r.POST("rule/upd-storage", func(c *gin.Context) {
		c.JSON(http.StatusOK, controller.Rule.ActionUpdStorage(c))
	})
	r.POST("rule/del-storage", func(c *gin.Context) {
		c.JSON(http.StatusOK, controller.Rule.ActionDelStorage(c))
	})
}
