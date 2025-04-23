package resource

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
)

var Header header

type header struct{}

func (h header) CheckManageToken(c *gin.Context) bool {
	token := c.GetHeader("manage-token")
	if token == "" {
		return false
	}
	return token == config.C.App.ManageToken
}

func (h header) CheckServiceSupportToken(c *gin.Context) bool {
	token := c.GetHeader("ss-token")
	if token == "" {
		return false
	}
	return token == config.C.App.ServiceSupportToken
}
