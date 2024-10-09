package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go-com/config"
	"go-com/core/mod"
	"go-com/core/tool"
	"strconv"
)

var Test test

type test struct {
}

func init() {
	config.AddRouterApi(test{}, &config.RouterApiWebList)
}

func (ctl test) ActExport(c *gin.Context) {
	var err error
	var param struct{}
	if err = c.ShouldBindJSON(&param); err != nil {
		c.JSON(400, tool.RespData(400, tool.ErrorStr(err), nil))
		return
	}

	query := make(map[string]interface{})
	query["select"] = []string{"id,device_ip,device_name,city,station,index,result,create_time"}

	_, err = mod.ExcelExport("稽核告警列表", []interface{}{"编号", "IP地址", "设备名称", "地市", "局站", "稽核项", "稽核时间", "稽核异常说明"},
		func(page int, pageSize int, isCount bool) mod.ExcelReadTable {
			var count int64
			var list []map[string]interface{}
			// 获取总数
			if isCount {
				query["select"] = []string{"count(*)"}
				query["limit"] = -1
				if len(list) > 0 {
					count = int64(list[0]["count"].(float64))
				}
			}

			query["select"] = []string{"id", "device_ip", "device_name", "city", "station", "index", "create_time", "result"}
			query["limit"] = pageSize
			query["order"] = []string{"id desc"}
			query["offset"] = (page - 1) * pageSize
			return mod.ExcelReadTable{
				Count: count,
				List:  list,
			}
		}, func(stream *excelize.StreamWriter, table mod.ExcelReadTable, rowNext *int) {
			var row []interface{}
			for _, item := range table.List.([]map[string]interface{}) { // 3 model类型
				// 4 表格字段
				row = []interface{}{item["id"], item["device_ip"], item["device_name"], item["city"], item["station"], "", tool.FormatStandardTime(item["create_time"].(string)), item["result"]}
				stream.SetRow("A"+strconv.Itoa(*rowNext), row)
				*rowNext++
			}
		}, c.Writer)
	if err != nil {
		c.JSON(500, tool.RespData(500, tool.ErrorStr(err), nil))
		return
	}

	return
}
