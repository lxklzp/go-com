package api

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/internal/api/controller"
	"go-com/internal/app"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func bind(r *gin.Engine) {
	controller.InitController()
	for _, routerApi := range config.RouterApiList {
		routerApi := routerApi
		r.POST(routerApi.Path, func(c *gin.Context) {
			c.JSON(http.StatusOK, routerApi.Action(c))
		})
	}

	// vmc代理
	urlVmc, _ := url.Parse("http://192.168.2.66:18080")
	app.ProxyVmc = httputil.NewSingleHostReverseProxy(urlVmc)
	r.POST("/index/api/webrtc", func(c *gin.Context) {
		app.ProxyVmc.ServeHTTP(c.Writer, c.Request)
	})
}
