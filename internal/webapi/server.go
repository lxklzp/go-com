package webapi

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"io"
	"net/http"
	"strings"
	"time"
)

var ServApi *http.Server

func Run() {
	// 表单验证错误信息的中文翻译
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		if err := zhTrans.RegisterDefaultTranslations(v, config.Trans); err != nil {
			logr.L.Fatal(err)
		}
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = logr.L.Out      // 设定日志
	gin.DefaultErrorWriter = logr.L.Out // 设定日志
	r := gin.New()

	r.MaxMultipartMemory = config.C.App.MaxMultipartMemory // 设置最大上传文件
	// 设置静态目录
	r.Static("/html", config.Root+"html")
	r.Static("/upload", config.C.App.PublicPath)

	r.Use(midGate, midRecovery) // 中间件

	bind(r) // 绑定接口
	// 启动
	ServApi = &http.Server{
		Addr:    config.C.App.WebApiAddr,
		Handler: r,
	}
	if err := ServApi.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logr.L.Fatal(err)
	}
}

// Shutdown 关闭
func Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := ServApi.Shutdown(ctx); err != nil {
		logr.L.Fatal(err)
	}
}

// 捕获异常中间件
func midRecovery(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			msg := tool.ErrorStack(err)
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
	if config.C.App.DebugMode {
		var header, body string
		// 请求header
		for k, v := range c.Request.Header {
			header += fmt.Sprintf("%s:%s\n", k, strings.Join(v, " "))
		}
		// 请求body
		buffer := config.BufPool.Get().(*bytes.Buffer)
		_, err := io.Copy(buffer, c.Request.Body)
		if err != nil {
			logr.L.Warn(err)
		}
		c.Request.Body = io.NopCloser(buffer)
		body = buffer.String()
		defer func() {
			buffer.Reset()
			config.BufPool.Put(buffer)
		}()
		// 写入日志文件
		logr.L.Infof("[request] %s\n%s %s %s %s %s\n--HEADER--\n%s--BODY--\n%s", c.Request.URL, c.Request.Method, c.Request.Proto, c.Request.Host, c.ClientIP(), c.RemoteIP(), header, body)
	}

	// 验证请求是否来自网关
	c.Next()
}
