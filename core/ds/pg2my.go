package ds

import (
	"go-com/core/logr"
	"go-com/core/my"
	"go-com/core/pg"
	"gorm.io/gorm"
	"os"
	"strings"
)

type Pg2My struct {
	sync *DataSync
}

// NewPg2My 创建postgresql到mysql的同步组件
func NewPg2My(planFilename string, srcDb *gorm.DB, srcCfg pg.Config, dstDb *gorm.DB, dstCfg my.Config) *Pg2My {
	// 初始化配置
	p2m := Pg2My{sync: NewSyncByPlan(planFilename)}
	if p2m.sync.Src.Table == nil {
		logr.L.Panic("请在计划中配置Src的Table")
	}
	if p2m.sync.Src.Schema == "" {
		logr.L.Panic("请在计划中配置Src的Schema")
	}
	if len(p2m.sync.Src.Cfg) == 0 {
		logr.L.Panic("请在计划中配置Src的Cfg")
	}
	if p2m.sync.Src.PageSize == 0 {
		logr.L.Panic("请在计划中配置Src的PageSize")
	}
	if len(p2m.sync.Dst.Cfg) == 0 {
		logr.L.Panic("请在计划中配置Dst的Cfg")
	}
	if p2m.sync.Dst.PageSize == 0 {
		logr.L.Panic("请在计划中配置Dst的PageSize")
	}
	// 初始化数据源连接
	p2m.sync.Src.Db = srcDb
	p2m.sync.Src.Cfg = map[string]string{"dbname": srcCfg.Dbname}
	// 初始化数据目标连接
	p2m.sync.Dst.Db = dstDb
	p2m.sync.Src.Cfg = map[string]string{"dbname": dstCfg.Dbname}

	// 设置源和目标的数据类型
	p2m.sync.DataTypeSrc = DataTypePg
	p2m.sync.DataTypeDst = DataTypeMy
	return &p2m
}

// Process 执行同步，追加的方式
func (p2m *Pg2My) Process() {
	processRelateDB(p2m.sync, p2m.CreateTable)
}

// CreateTable 根据源表，创建相同表名、相同结构的目标表，如果目标表已存在，则跳过
func (p2m *Pg2My) CreateTable() {
	logr.L.Info("正在创建表...")
	// 1 得到查询的tableName
	tableName := generateTableNameString(&p2m.sync.Src, p2m.sync.DataTypeSrc)
	// 2 生成model代码文件
	GenModel(p2m.sync.Src.Db, tableName, "runtime/dao", p2m.sync.Dst.Db)
	// 3 生成创建表的代码文件
	if tableNeedCreate := myGenerateCreateTableCode(&p2m.sync.Dst, tableName); len(tableNeedCreate) > 0 {
		logr.L.Infof("运行以下指令，用于创建表[%s]：\ngo run create_db_table.go", strings.Join(tableNeedCreate, ","))
		os.Exit(0)
	}
}
