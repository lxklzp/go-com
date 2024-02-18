package api

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/global"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var reqBufPool *sync.Pool

func Run() {
	reqBufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4096))
		},
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = global.Log.Out      // 设定日志
	gin.DefaultErrorWriter = global.Log.Out // 设定日志
	r := gin.New()

	r.Use(midGate, midRecovery) // 中间件

	bind(r) // 绑定接口

	// 启动
	global.ServeApi = &http.Server{
		Addr:    config.C.App.ApiAddr,
		Handler: r,
	}
	if err := global.ServeApi.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		global.Log.Fatal(err)
	}
}

// Shutdown 关闭
func Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := global.ServeApi.Shutdown(ctx); err != nil {
		global.Log.Fatal(err)
	}
}

// 捕获异常中间件
func midRecovery(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			msg := global.ErrorStack(err)
			if config.C.App.DebugMode {
				c.JSON(http.StatusOK, map[string]interface{}{"code": 500, "message": msg})
			} else {
				c.JSON(http.StatusOK, map[string]interface{}{"code": 500, "message": "服务器异常"})
			}
		}
	}()
	c.Next()
}

// 请求与响应处理中间件
func midGate(c *gin.Context) {
	var header, body string
	// 请求header
	for k, v := range c.Request.Header {
		header += fmt.Sprintf("%s:%s\n", k, strings.Join(v, " "))
	}
	// 请求body
	buffer := reqBufPool.Get().(*bytes.Buffer)
	_, err := io.Copy(buffer, c.Request.Body)
	if err != nil {
		global.Log.Warn(err)
	}
	c.Request.Body = io.NopCloser(buffer)
	body = buffer.String()
	defer func() {
		buffer.Reset()
		reqBufPool.Put(buffer)
	}()
	// 写入日志文件
	global.Log.Infof("[request] %s\n%s %s %s %s %s\n--HEADER--\n%s--BODY--\n%s", c.Request.URL, c.Request.Method, c.Request.Proto, c.Request.Host, c.ClientIP(), c.RemoteIP(), header, body)

	// 验证请求是否来自网关
	c.Next()
}
