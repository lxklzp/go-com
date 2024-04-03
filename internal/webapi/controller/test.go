package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/tool"
	"go-com/internal/app"
	"strings"
)

type test struct {
}

func init() {
	config.AddRouterApi(test{}, &config.RouterApiWebList)
}

type ApiItem struct {
	Name   string `json:"name"`
	Method string `json:"method"`
}

// ActionRefreshApi 刷新api到认证和鉴权中心
func (ctl test) ActionRefreshApi(c *gin.Context) interface{} {
	reqData := map[string]interface{}{
		"platform_id": 3,
		"api_prefix":  "api/robot/web",
	}

	routes := app.ServeApi.Handler.(*gin.Engine).Routes()
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
			return tool.RespData(500, data.Message, nil)
		}
	} else {
		return tool.RespData(500, err.Error(), nil)
	}

	return tool.RespData(200, "", data.Data)
}
