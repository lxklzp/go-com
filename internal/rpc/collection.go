package rpc

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/tool"
)

var Collection = collection{}

type collection struct {
}

type acquisitionRespData[T interface{} | DeviceRaw] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data"`
}

func CollectionPost[T interface{} | DeviceRaw](url string, param interface{}, respData acquisitionRespData[T]) (T, error) {
	bsBuf := config.BufPool.Get().(*bytes.Buffer)
	defer func() {
		bsBuf.Reset()
		config.BufPool.Put(bsBuf)
	}()

	encoder := json.NewEncoder(bsBuf)
	encoder.SetEscapeHTML(false)
	encoder.Encode(param)
	paramJson := bsBuf.Bytes()

	header := map[string]string{}
	var result []byte
	var err error
	if result, err = tool.Post(url, paramJson, header); err == nil {
		json.Unmarshal(result, &respData)
		if respData.Code != 200 {
			return respData.Data, errors.New(respData.Msg)
		}
		return respData.Data, nil
	}
	return respData.Data, err
}

func (r *collection) Inspection(param map[string]interface{}) interface{} {
	if result, err := CollectionPost(config.C.App.GatewayAddr+"/api/newcitynet/inspection", param, acquisitionRespData[interface{}]{}); err != nil {
		logr.L.Error(err)
		return nil
	} else {
		return result
	}
}

type Slot struct {
	Position string  `json:"position"`
	SlotType string  `json:"slot_type"`
	Status   bool    `json:"status"`
	Boards   []Board `json:"boards"`
}

type Chassis struct {
	Position      string `json:"position"`
	ShelfType     string `json:"shelf_type"`
	ChassisStatus bool   `json:"chassis_status"`
	Slot          struct {
		Slots []Slot `json:"slots"`
	} `json:"slot"`
}

type SubSlot struct {
	Position string `json:"position"`
	SlotType string `json:"slot_type"`
	Cards    []Card `json:"cards"`
}

type Board struct {
	Position         string `json:"position"`
	BoardSerial      string `json:"board_serial"`
	BoardName        string `json:"board_name"`
	BoardStatus      string `json:"board_status"`
	BoardBomId       string `json:"board_bom_id"`
	HardwareVersion  string `json:"hardware_version"`
	HardwareModel    string `json:"hardware_model"`
	ManufacturerDate string `json:"manufacturer_date"`
	ManufacturerName string `json:"manufacturer_name"`
	Memory           string `json:"memory"`
	BoardDescription string `json:"board_description"`
	BoardType        string `json:"board_type"`
	SubSlot          struct {
		SubSlots []SubSlot `json:"sub_slots"`
	} `json:"sub_slot"`
	AlarmLevel int `json:"alarm_level"`
}

type Port map[string]interface{}

type Card struct {
	Position   string `json:"position"`
	CardStatus string `json:"card_status"`
	CardType   string `json:"card_type"`
	Port       struct {
		Ports []Port `json:"ports"`
	} `json:"port"`
}

// DeviceRaw 原始设备信息
type DeviceRaw struct {
	Chassis []Chassis                `json:"chassis"`
	Boards  []Board                  `json:"boards"`
	Cards   []Card                   `json:"cards"`
	Fans    []map[string]interface{} `json:"fans"`
	Powers  []map[string]interface{} `json:"powers"`
}

// DeviceDst 树状结构的设备信息
type DeviceDst struct {
	Chassis []Chassis                `json:"chassis"`
	Fans    []map[string]interface{} `json:"fans"`
	Powers  []map[string]interface{} `json:"powers"`
}
