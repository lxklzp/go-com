package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/mod"
	"go-com/core/pg"
	"go-com/internal/app"
	"go-com/internal/model"
	"strconv"
)

func main() {
	config.Load()
	logr.InitLog("test")
	app.Pg = pg.NewDb(pg.Config{Postgresql: config.C.Postgresql})

	// 1 名称 标题
	fmt.Println(mod.ExcelExport("用户列表", []interface{}{"名称", "年龄", "ID", "用户ID", "创建时间"},
		func(page int, pageSize int, isCount bool) mod.ExcelReadTable {
			// 2 总数 列表
			var list []model.Hehe
			query := app.Pg.Model(model.Hehe{})
			var count int64
			if isCount {
				query.Count(&count)
			}
			query.Order("id desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&list)

			return mod.ExcelReadTable{count, list}
		}, func(stream *excelize.StreamWriter, table mod.ExcelReadTable, rowNext *int) {
			var row []interface{}
			for _, item := range table.List.([]model.Hehe) { // 3 model类型
				// 4 表格字段
				row = []interface{}{item.Name, item.Age, item.ID, item.UserID, item.CreateTime.String()}
				stream.SetRow("A"+strconv.Itoa(*rowNext), row)
				*rowNext++
			}
		}))
}
