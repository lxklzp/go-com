package rpc

import (
	"encoding/json"
	"github.com/pkg/errors"
	"go-com/config"
	"go-com/global"
	"net/http"
)

var Auth = auth{
	api: map[string]string{
		"current-user-info": "/auth/auth-api/current-user-info",
		"get-user-info":     "/auth/auth-api/get-user-info",
	},
}

type auth struct {
	api map[string]string
}

func (r *auth) post(api string, param map[string]interface{}, req *http.Request) (global.ResponseData, error) {
	paramJson, _ := json.Marshal(param)
	header := map[string]string{
		"Authorization": req.Header.Get("Authorization"),
		"Api-Platform":  req.Header.Get("Api-Platform"),
	}
	var data global.ResponseData
	var result []byte
	var err error
	if result, err = global.Post(config.C.App.GatewayAddr+r.api[api], paramJson, header); err == nil {
		json.Unmarshal(result, &data)
		if data.Code != 200 {
			return data, errors.New(data.Message)
		}
		return data, nil
	}
	return data, err
}

func (r *auth) CurrentUserId(req *http.Request) int32 {
	data, err := r.post("current-user-info", nil, req)
	if err != nil {
		return 0
	}
	return int32(data.Data.(map[string]interface{})["id"].(float64))
}

func (r *auth) GetUserInfo(req *http.Request, ids []int32) map[string]map[string]string {
	data, err := r.post("get-user-info", map[string]interface{}{"ids": ids}, req)
	if err != nil {
		return nil
	}

	list := make(map[string]map[string]string)
	raw := data.Data.([]interface{})
	for _, v := range raw {
		item := make(map[string]string)
		for k, v := range v.(map[string]interface{}) {
			item[k] = global.InterfaceToString(v)
		}
		list[item["id"]] = item
	}
	return list
}
