package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/api"
	"strings"
)

type ApiItem struct {
	Name   string `json:"name"`
	Method string `json:"method"`
}

func main() {
	config.Load()
	logr.InitLog("refresh_api")

	reqData := map[string]interface{}{
		"platform_id": 173,
		"api_prefix":  "api/go-com",
	}

	//api.TestServApi()
	routes := api.ServApi.Handler.(*gin.Engine).Routes()
	var apiList []ApiItem
	for _, route := range routes {
		// 过滤掉不需要提交的接口：静态文件路由、测试接口
		if strings.HasPrefix(route.Path, "/uploads/") || strings.HasPrefix(route.Path, "/public/") || strings.HasPrefix(route.Path, "/test/") {
			continue
		}
		apiList = append(apiList, ApiItem{
			Name:   route.Path,
			Method: route.Method,
		})
	}
	reqData["list"] = apiList

	paramJson, _ := json.Marshal(reqData)
	var data tool.ResponseData
	var result []byte
	var err error
	// 调用认证和鉴权中心的刷新应用接口列表
	url := config.C.App.GatewayAddr + "/api/auth/test/refresh-api"
	if result, err = tool.Post(url, paramJson, nil); err == nil {
		json.Unmarshal(result, &data)
		if data.Code != 200 {
			logr.L.Error(data.Message)
		}
	} else {
		logr.L.Error(err)
	}
}
