package rpc

import (
	"encoding/json"
	"github.com/pkg/errors"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
	"go-com/internal/model"
)

var App = app{}

type app struct {
}

func (r *app) Post(url string, param interface{}) (interface{}, error) {
	paramJson, _ := json.Marshal(param)
	header := map[string]string{
		"Gateway-Token": config.C.App.GatewayToken,
	}
	var data tool.ResponseData
	var result []byte
	var err error
	if result, err = tool.Post(url, paramJson, header); err == nil {
		json.Unmarshal(result, &data)
		if data.Code != 200 {
			return nil, errors.New(data.Message)
		}
		return data.Data, nil
	}
	return nil, err
}

func (r *app) PowerGroupAdd(m model.RuleStorage) bool {
	if _, err := r.Post("http://127.0.0.1:8081/merge/power-group-add", m); err != nil {
		logr.L.Error(err)
		return false
	}
	return true
}
