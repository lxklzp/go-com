package webApi

import (
	"github.com/gin-gonic/gin"
	"go-com/internal/webapi/controller"
	"net/http"
)

func bind(r *gin.Engine) {
	r.POST("rule-storage/add", func(c *gin.Context) {
		c.JSON(http.StatusOK, controller.RuleStorage.ActionAdd(c))
	})
	r.POST("rule-storage/upd", func(c *gin.Context) {
		c.JSON(http.StatusOK, controller.RuleStorage.ActionUpd(c))
	})
	r.POST("rule-storage/list", func(c *gin.Context) {
		c.JSON(http.StatusOK, controller.RuleStorage.ActionList(c))
	})
	r.POST("rule-storage/export", func(c *gin.Context) {
		c.JSON(http.StatusOK, controller.RuleStorage.ActionExport(c))
	})
	r.POST("rule-storage/del", func(c *gin.Context) {
		c.JSON(http.StatusOK, controller.RuleStorage.ActionDel(c))
	})
}
