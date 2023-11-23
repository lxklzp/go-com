package rpc

import (
	"encoding/json"
	"github.com/pkg/errors"
	"go-com/config"
	"go-com/global"
	"go-com/internal/model"
	"go-com/lib/service"
)

var App = app{}

type app struct {
}

func (r *app) Post(url string, param interface{}) (interface{}, error) {
	paramJson, _ := json.Marshal(param)
	header := map[string]string{
		"Gateway-Token": config.C.App.GatewayToken,
	}
	var data global.ResponseData
	var result []byte
	var err error
	if result, err = global.Post(url, paramJson, header); err == nil {
		json.Unmarshal(result, &data)
		if data.Code != 200 {
			return nil, errors.New(data.Message)
		}
		return data.Data, nil
	}
	return nil, err
}

func (r *app) PowerGroupAdd(m model.RuleStorage) bool {
	addr := service.SD.DiscoveryByConsistentHash(service.SDMergePrefix, m.ID)
	if addr == "" {
		return false
	}
	if _, err := r.Post(addr+"merge/power-group-add", m); err != nil {
		global.Log.Error(err)
		return false
	}
	return true
}
