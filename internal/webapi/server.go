package webapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var reqBufPool *sync.Pool

var ServApi *http.Server

func Run() {
	reqBufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4096))
		},
	}

	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = logr.L.Out      // 设定日志
	gin.DefaultErrorWriter = logr.L.Out // 设定日志
	r := gin.New()

	pprof.RouteRegister(r, config.WebApiPrefix+"pprof") // pprof

	r.MaxMultipartMemory = config.C.App.MaxMultipartMemory // 设置最大上传文件

	r.Use(midGate, midRecovery) // 中间件

	r.Static(config.WebApiPrefix+"public", config.C.App.PublicPath)

	bind(r) // 绑定接口

	// 启动
	ServApi = &http.Server{
		Addr:    config.C.App.WebAddr,
		Handler: r,
	}
	if err := ServApi.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
		buffer := reqBufPool.Get().(*bytes.Buffer)
		_, err := io.Copy(buffer, c.Request.Body)
		if err != nil {
			logr.L.Warn(err)
		}
		c.Request.Body = io.NopCloser(buffer)
		body = buffer.String()
		defer func() {
			buffer.Reset()
			reqBufPool.Put(buffer)
		}()
		// 写入日志文件
		logr.L.Infof("[request] %s\n%s %s %s %s %s\n--HEADER--\n%s--BODY--\n%s", c.Request.URL, c.Request.Method, c.Request.Proto, c.Request.Host, c.ClientIP(), c.RemoteIP(), header, body)
	}

	if config.C.App.GatewayToken != "" {
		// 验证请求是否来自网关
		if len(c.Request.Header["Gateway-Token"]) > 0 && c.Request.Header["Gateway-Token"][0] == config.C.App.GatewayToken {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, tool.RespData(403, "非法访问。", nil))
		}
	}
	c.Next()
}
