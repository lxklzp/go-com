package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"go-com/config"
	"go-com/core/mod"
	"go-com/core/tool"
	"go-com/internal/app"
	"go-com/internal/model"
	"go-com/internal/system"
	"gorm.io/gorm"
	"strconv"
	"time"
)

/* 列表和导出接口示例 */

type listExport struct{}

func init() {
	config.AddRouterApi(listExport{}, &config.RouterApiList)
}

type DigitalLifeIndexReqData struct {
	Id          int    `json:"id"`
	IndicatorID int    `json:"indicatorId"`
	ProductID   int    `json:"product_id"`
	Type        string `json:"type"`
	mod.Base
}

func (ctl listExport) queryDigitalLifeIndexList(param DigitalLifeIndexReqData) *gorm.DB {
	query := app.Pg.Model(model.DigitalLifeIndex{})
	if param.ProductID != 0 {
		query.Where("product_id=?", param.ProductID)
	}
	return query
}

func (ctl listExport) formatDigitalLifeIndexList(mList []model.DigitalLifeIndex) {
	indexMap := model.CodeMap("digital_life_index")
	productMap := make(map[string]string)
	var mCode []model.Code
	app.Pg.Select("p_key,content").Where("type='digital_life_index'").Group("p_key,content").Find(&mCode)
	for _, item := range mCode {
		productMap[item.PKey] = item.Content
	}
	for k := range mList {
		mList[k].IndicatorName = indexMap[strconv.Itoa(mList[k].IndicatorID)]
		mList[k].ProductName = productMap[strconv.Itoa(mList[k].ProductID)]
		mList[k].TimeName = system.ShiLian.DigitalLifeIndexTimeType[mList[k].TimeType]
	}
}

// ActionDigitalLifeIndexList 列表
func (ctl listExport) ActionDigitalLifeIndexList(c *gin.Context) interface{} {
	var err error
	var param DigitalLifeIndexReqData
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, tool.ErrorStr(err), nil)
	}
	param.Base.Validate()

	var mList []model.DigitalLifeIndex
	query := ctl.queryDigitalLifeIndexList(param)

	var count int64
	query.Count(&count)

	query.Select("*").Limit(param.PageSize).Offset((param.Page - 1) * param.PageSize).Order("id desc").Find(&mList)

	ctl.formatDigitalLifeIndexList(mList)

	indexMap := model.CodeMapFull("digital_life_index")
	productIndexMap := make(map[string][]model.Code)
	for _, im := range indexMap {
		productIndexMap[im.PKey] = append(productIndexMap[im.PKey], im)
	}
	productMap := make(map[string]string)
	var mCode []model.Code
	mCodeWhere := fmt.Sprintf(`type='digital_life_index' and comment='%s'`, param.Type)
	app.Pg.Select("p_key,content").Where(mCodeWhere).Group("p_key,content").Find(&mCode)
	for _, item := range mCode {
		productMap[item.PKey] = item.Content
	}
	return tool.RespData(200, "", map[string]interface{}{
		"indicator": productIndexMap,
		"product":   productMap,
		"list":      mList,
		"count":     count,
	})
}

// ActionDigitalLifeIndexListExport 导出excel，前端传标题和字段
func (ctl listExport) ActionDigitalLifeIndexListExport(c *gin.Context) interface{} {
	var err error
	var param DigitalLifeIndexReqData
	if err = c.ShouldBindJSON(&param); err != nil {
		return tool.RespData(400, tool.ErrorStr(err), nil)
	}

	var mDownloadId int
	if mDownloadId, err = system.DownloadBefore(param.ExportTitle, 0); err != nil {
		return tool.RespData(400, err.Error(), nil)
	}

	go func() {
		path, err := mod.ExcelExport(param.ExportTitle, param.ExportHeader,
			func(page int, pageSize int, isCount bool) mod.ExcelReadTable {
				var list []model.DigitalLifeIndex
				query := ctl.queryDigitalLifeIndexList(param)
				var count int64
				if isCount {
					query.Count(&count)
				}
				query.Select("*").Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list)

				ctl.formatDigitalLifeIndexList(list)

				return mod.ExcelReadTable{count, list}
			}, func(stream *excelize.StreamWriter, table mod.ExcelReadTable, rowNext *int) {
				var row []interface{}
				for _, item := range table.List.([]model.DigitalLifeIndex) {
					row = row[:0]
					mMap, _ := tool.StructToMap(item, "json")
					for _, k := range param.ExportField {
						switch mMap[k].(type) {
						case config.Timestamp:
							mMap[k] = time.Time(mMap[k].(config.Timestamp)).Format(config.DateTimeFormatter)
						}
						row = append(row, mMap[k])
					}
					stream.SetRow("A"+strconv.Itoa(*rowNext), row)
					*rowNext++
				}
			}, nil)

		system.DownloadAfter(mDownloadId, path, err)
	}()

	return tool.RespData(200, "", map[string]interface{}{
		"download_id": mDownloadId,
	})
}
