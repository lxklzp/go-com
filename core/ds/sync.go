package ds

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/my"
	"go-com/core/pg"
	"go-com/core/tool"
	"strconv"
	"strings"
	"time"
)

// NewSyncByPlan 根据计划配置实例化一个同步组件
func NewSyncByPlan(plan interface{}) *DataSync {
	var err error
	switch plan.(type) {
	case string:
		// 载入计划
		v := viper.New()
		path := config.Root + "data/"
		file := path + plan.(string)
		v.SetConfigFile(file)
		viper.AddConfigPath(path)
		if err = v.ReadInConfig(); err != nil {
			logr.L.Panic(err)
		}
		var sync DataSync
		if err = v.Unmarshal(&sync); err != nil {
			logr.L.Panic(err)
		}
		return &sync
	case []byte:
		v := viper.New()
		v.SetConfigType("yaml")
		if err = v.ReadConfig(bytes.NewBuffer(plan.([]byte))); err != nil {
			logr.L.Panic(err)
		}
		var sync DataSync
		if err = v.Unmarshal(&sync); err != nil {
			logr.L.Panic(err)
		}
		return &sync
	case *DataSync:
		return plan.(*DataSync)
	default:
		logr.L.Panic("计划有误")
	}
	return nil
}

/*---------- 关系型数据库基于GORM实现的相关方法 开始 ----------*/

// 执行同步，追加的方式
func processRelateDB(sync *DataSync, createTable func()) {
	begin := time.Now()
	logr.L.Infof("开始时间：%s，同步中...", begin.Format(config.DateTimeFormatter))
	switch sync.Src.Table.(type) {
	case string:
		if sync.Src.Table == ForDataBase {
			createTable()
			var rows []map[string]interface{}
			switch sync.DataTypeSrc {
			case DataTypePg:
				rows = pg.GetDbTables(sync.Src.Db, sync.Src.Schema)
			case DataTypeMy:
				rows = my.GetDbTables(sync.Src.Db, sync.Src.Cfg["dbname"])
			}
			for _, row := range rows {
				if sync.DataTypeSrc == DataTypePg {
					syncTable(sync, fmt.Sprintf(`"%s"."%s"`, sync.Src.Schema, row["table_name"].(string)))
				} else {
					syncTable(sync, row["table_name"].(string))
				}
			}
		} else {
			if sync.DataTypeSrc == DataTypePg {
				syncSingleTable(sync, fmt.Sprintf(`"%s"."%s"`, sync.Src.Schema, sync.Src.Table.(string)))
			} else {
				syncSingleTable(sync, sync.Src.Table.(string))
			}
		}
	case []interface{}:
		createTable()
		for _, tableName := range sync.Src.Table.([]interface{}) {
			if sync.DataTypeSrc == DataTypePg {
				syncTable(sync, fmt.Sprintf(`"%s"."%s"`, sync.Src.Schema, tableName.(string)))
			} else {
				syncTable(sync, tableName.(string))
			}
		}
	}
	logr.L.Infof("结束时间：%s，共计耗时：%s，同步完成。", begin.Format(config.DateTimeFormatter), time.Since(begin))
}

// 多表同步时，重置缓存数据
func syncTable(sync *DataSync, tableNameSrc string) {
	syncReset(sync, tableNameSrc)
	syncSingleTable(sync, tableNameSrc)
}

// 多表同步时，重置缓存数据
func syncReset(sync *DataSync, tableNameSrc string) {
	// 重置源缓存数据
	sync.Src.Table = tableNameSrc
	sync.Src.Key = nil
	sync.Src.Field = nil
	sync.Src.Page = 0
	sync.Src.Count = 0
	sync.Src.DataCacheMap = nil
	sync.Src.DataCacheArray = nil
	sync.Src.PrimaryValueList = nil
	sync.Src.FieldSql = ""
	sync.Src.KeySql = ""

	// 重置目标缓存数据
	sync.Dst.Table = ""
	sync.Dst.Field = nil
	sync.Dst.Step = 0
	sync.Dst.StepPeriod = 0
	sync.Dst.FieldSql = ""
	sync.Dst.UpdFieldSql = ""
	sync.Dst.ValueSql = nil
	sync.Dst.Sql = nil
}

func syncSingleTable(sync *DataSync, tableNameSrc string) {
	// 源
	switch sync.DataTypeSrc {
	case DataTypePg:
		pgBeforeReadSingleTable(&sync.Src, tableNameSrc)
	case DataTypeMy:
		myBeforeReadSingleTable(&sync.Src, tableNameSrc)
	}
	logr.L.Infof("%s表待同步数据约为：%d条。", tableNameSrc, sync.Src.Count)
	readSingleTable(sync, tableNameSrc)
	// 目标
	dataCacheMapLen := len(sync.Src.DataCacheMap)
	if dataCacheMapLen > 0 && sync.Src.Page > 0 {
		beforeWriteSingleTable(sync, tableNameSrc)

		sync.Dst.SlicePage.Before(int64(dataCacheMapLen), int64(sync.Dst.PageSize))
		for sync.Dst.SlicePage.Next() {
			writeSingleTable(sync, sync.Src.DataCacheMap[sync.Dst.SlicePage.From:sync.Dst.SlicePage.To])
		}

		// 分页读写
		for {
			readSingleTable(sync, tableNameSrc)
			dataCacheMapLen = len(sync.Src.DataCacheMap)
			if dataCacheMapLen == 0 {
				break
			}
			sync.Dst.SlicePage.Before(int64(dataCacheMapLen), int64(sync.Dst.PageSize))
			for sync.Dst.SlicePage.Next() {
				writeSingleTable(sync, sync.Src.DataCacheMap[sync.Dst.SlicePage.From:sync.Dst.SlicePage.To])
			}
		}
	}

	// 执行收尾sql
	if sync.Dst.EndSql != "" {
		if err := sync.Dst.Db.Exec(sync.Dst.EndSql).Error; err != nil {
			logr.L.Panic(err)
		}
	}
}

// 分页读取表数据 postgresql/mysql
func readSingleTable(sync *DataSync, tableName string) {
	var err error
	sync.Src.DataCacheMap = sync.Src.DataCacheMap[:0]
	if sync.Src.KeySql == "" {
		// 无主键，查询所有数据
		if err = sync.Src.Db.Table(tableName).Select(sync.Src.FieldSql).Where(sync.Src.Where).Find(&sync.Src.DataCacheMap).Error; err != nil {
			logr.L.Panic(err)
		}
	} else if !strings.Contains(sync.Src.KeySql, ",") {
		// 单主键，先分页查询主键，再根据主键查询字段
		sync.Src.Db.Table(tableName).Select(sync.Src.KeySql).Where(sync.Src.Where).Limit(sync.Src.PageSize).Offset(sync.Src.Page * sync.Src.PageSize).Order(sync.Src.KeySql).Find(&sync.Src.DataCacheMap)
		if len(sync.Src.DataCacheMap) == 0 {
			return
		}
		sync.Src.PrimaryValueList = sync.Src.PrimaryValueList[:0]
		key := strings.Trim(strings.Trim(sync.Src.KeySql, `"`), "`")
		for _, dc := range sync.Src.DataCacheMap {
			sync.Src.PrimaryValueList = append(sync.Src.PrimaryValueList, dc[key])
		}
		sync.Src.DataCacheMap = sync.Src.DataCacheMap[:0]
		if err = sync.Src.Db.Table(tableName).Select(sync.Src.FieldSql).Where(sync.Src.KeySql+" in ?", sync.Src.PrimaryValueList).Find(&sync.Src.DataCacheMap).Error; err != nil {
			logr.L.Panic(err)
		}
		sync.Src.Page++
	} else {
		// 多主键，直接分页查询字段
		if err = sync.Src.Db.Table(tableName).Select(sync.Src.FieldSql).Where(sync.Src.Where).Limit(sync.Src.PageSize).Offset(sync.Src.Page * sync.Src.PageSize).Order(sync.Src.KeySql).Find(&sync.Src.DataCacheMap).Error; err != nil {
			logr.L.Panic(err)
		}
		sync.Src.Page++
	}
}

// 生成一行写入值sql postgresql/mysql
func generateValueSql(sync *DataSync, dc interface{}) {
	var vStr string
	sync.Dst.ValueSql = sync.Dst.ValueSql[:0]
	switch dc.(type) {
	case map[string]interface{}:
		dc := dc.(map[string]interface{})
		for i := range sync.Src.Field {
			vStr = interfaceToString(dc[sync.Src.Field[i]])
			if vStr == NullStr {
				sync.Dst.ValueSql = append(sync.Dst.ValueSql, []byte("null,")...)
			} else {
				sync.Dst.ValueSql = append(sync.Dst.ValueSql, []byte("'"+vStr+"',")...)
			}
		}
	case []string:
		dc := dc.([]string)
		dcCount := len(dc)
		for _, i := range sync.Src.Column {
			if dcCount <= i {
				sync.Dst.ValueSql = append(sync.Dst.ValueSql, []byte("null,")...)
			} else {
				sync.Dst.ValueSql = append(sync.Dst.ValueSql, []byte("'"+dc[i]+"',")...)
			}
		}
	}
	// 设置默认值
	for _, v := range sync.Dst.DefaultValue {
		vStr = formatDefaultValue(interfaceToString(v))
		sync.Dst.ValueSql = append(sync.Dst.ValueSql, []byte("'"+vStr+"',")...)
	}
	sync.Dst.ValueSql = []byte("(" + strings.TrimSuffix(string(sync.Dst.ValueSql), ",") + "),")
}

// 写入表数据前的准备工作 postgresql/mysql
func beforeWriteSingleTable(sync *DataSync, tableNameSrc string) {
	if len(sync.Dst.Field) == 0 {
		sync.Dst.Field = sync.Src.Field
	}

	if sync.Dst.Table == "" {
		if sync.DataTypeSrc == DataTypePg {
			sync.Dst.Table = strings.Trim(strings.Split(tableNameSrc, ".")[1], `"`)
		} else {
			sync.Dst.Table = tableNameSrc
		}
	}
	switch sync.DataTypeDst {
	case DataTypePg:
		sync.Dst.FieldSql = pg.GenerateFieldSql(append(sync.Dst.Field, sync.Dst.DefaultField...))
		sync.Dst.UpdFieldSql = pg.GenerateFieldSql(sync.Dst.UpdField)
		sync.Dst.Table = fmt.Sprintf(`"%s"."%s"`, sync.Dst.Schema, sync.Dst.Table)
		// 是否清空表
		if sync.Dst.Truncate {
			//if err := sync.Dst.Db.Exec("").Error; err != nil {
			//	logr.L.Panic(err)
			//}
		}
	case DataTypeMy:
		sync.Dst.FieldSql = my.GenerateFieldSql(append(sync.Dst.Field, sync.Dst.DefaultField...))
		sync.Dst.UpdFieldSql = my.GenerateUpdFieldSql(sync.Dst.UpdField)
		// 是否清空表
		if sync.Dst.Truncate {
			if err := sync.Dst.Db.Exec(fmt.Sprintf("truncate table %s", sync.Dst.Table)).Error; err != nil {
				logr.L.Panic(err)
			}
		}
	}

	// 给一行值sql初始化内存，1kb
	sync.Dst.ValueSql = make([]byte, 0, 1024)
	// 根据第一行数据调整一行值sql的内存
	if sync.Src.DataCacheMap != nil {
		generateValueSql(sync, sync.Src.DataCacheMap[0])
	} else {
		generateValueSql(sync, sync.Src.DataCacheArray[0])
	}
	// 给sql初始化内存
	sync.Dst.Sql = make([]byte, 0, len(sync.Dst.ValueSql)*sync.Dst.PageSize+1024)
}

// 从读缓存中写入表数据 postgresql/mysql
func writeSingleTable[T map[string]interface{} | []string](sync *DataSync, list []T) {
	sync.Dst.Sql = sync.Dst.Sql[:0]
	// 遍历读缓存
	for _, dc := range list {
		generateValueSql(sync, dc)
		sync.Dst.Sql = append(sync.Dst.Sql, sync.Dst.ValueSql...)
	}
	sync.Dst.Step += int64(len(list))

	switch sync.DataTypeDst {
	case DataTypePg:
		sync.Dst.Sql = []byte("INSERT INTO " + sync.Dst.Table + " (" + sync.Dst.FieldSql + ") VALUES" + strings.TrimSuffix(string(sync.Dst.Sql), ","))
		if sync.Dst.UpdFieldSql != "" {
			sync.Dst.Sql = append(sync.Dst.Sql, []byte(" ON CONFLICT ("+sync.Dst.UpdFieldSql+") DO NOTHING")...)
		}
	case DataTypeMy:
		sync.Dst.Sql = []byte("INSERT INTO " + sync.Dst.Table + " (" + sync.Dst.FieldSql + ") VALUES" + strings.TrimSuffix(string(sync.Dst.Sql), ","))
		if sync.Dst.UpdFieldSql != "" {
			sync.Dst.Sql = append(sync.Dst.Sql, []byte(" ON DUPLICATE KEY UPDATE "+sync.Dst.UpdFieldSql)...)
		}
	}

	if err := sync.Dst.Db.Exec(string(sync.Dst.Sql)).Error; err != nil {
		logr.L.Panic(err)
	}
	logr.L.Infof("成功写入至第%d条（包含已存在的数据）", sync.Dst.Step)
}

// 根据计划配置，生成表名数组
func generateTableNameString(src *Src, dataType string) []string {
	var tableName []string
	switch src.Table.(type) {
	case string:
		if src.Table == ForDataBase {
			var rows []map[string]interface{}
			switch dataType {
			case DataTypePg:
				rows = pg.GetDbTables(src.Db, src.Schema)
			case DataTypeMy:
				rows = my.GetDbTables(src.Db, src.Cfg["dbname"])
			}
			for _, row := range rows {
				tableName = append(tableName, row["table_name"].(string))
			}
		} else {
			tableName = []string{src.Table.(string)}
		}
	case []interface{}:
		for _, t := range src.Table.([]interface{}) {
			switch dataType {
			case DataTypePg:
				tableName = append(tableName, src.Schema+"."+t.(string))
			case DataTypeMy:
				tableName = append(tableName, t.(string))
			}
		}
	}
	return tableName
}

// SyncReadScroll 游标读取数据，scrollId为空，表示第一步，不为空，表示后一步
func SyncReadScroll(scrollId string, sync *DataSync) (ReadScrollData, error) {
	tableName := sync.Src.Table.(string)
	readScrollData := ReadScrollData{ScrollId: NullStr, List: []map[string]interface{}{}, Count: 0}
	if scrollId != "" {
		if scrollId == NullStr {
			return readScrollData, nil
		}
		// 读取游标
		if syncAny, ok := dataSyncReadCache.Load(scrollId); !ok {
			return readScrollData, errors.New("scroll_id不存在")
		} else {
			// 初始化游标
			sync = syncAny.(*DataSync)
			sync.Src.DataCacheMap = make([]map[string]interface{}, 0, sync.Src.PageSize)
			sync.Src.PrimaryValueList = make([]interface{}, 0, sync.Src.PageSize)
		}
	} else {
		if tableName == "" {
			return readScrollData, errors.New("table_name不能为空")
		}
		// 初始化游标
		switch sync.DataTypeSrc {
		case DataTypePg:
			pgBeforeReadSingleTable(&sync.Src, tableName)
		case DataTypeMy:
			myBeforeReadSingleTable(&sync.Src, tableName)
		}
	}

	// 读取数据
	readSingleTable(sync, tableName)
	length := len(sync.Src.DataCacheMap)
	readScrollData.List = make([]map[string]interface{}, 0, length)
	if length > 0 {
		for _, m := range sync.Src.DataCacheMap {
			readScrollData.List = append(readScrollData.List, m)
		}
	}
	readScrollData.Count = sync.Src.Count

	if sync.Src.KeySql != "" && length != 0 {
		// 缓存游标
		sync.Src.DataCacheMap = nil
		sync.Src.PrimaryValueList = nil
		sync.Dst.Step = time.Now().Unix() + SyncReadScrollTimeout // 缓存20分钟
		readScrollData.ScrollId = strconv.Itoa(int(tool.SnowflakeComm.GetId()))
		dataSyncReadCache.Store(readScrollData.ScrollId, sync)
	}

	return readScrollData, nil
}

// SyncReadScrollClear 游标缓存数据清理
func SyncReadScrollClear() {
	dataSyncReadCache.Range(func(key, value any) bool {
		sync := value.(*DataSync)
		if sync.Dst.Step <= time.Now().Unix() {
			dataSyncReadCache.Delete(key)
		}
		return true
	})
}

/*---------- 关系型数据库基于GORM实现的相关方法 结束 ----------*/
