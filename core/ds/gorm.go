package ds

import (
	"go-com/config"
	"go-com/core/tool"
	"gorm.io/gen"
	"gorm.io/gorm"
	"strings"
)

type defaultPgModel struct {
}

func (m *defaultPgModel) TableName() string {
	return strings.Split("@@table", ".")[1]
}

func GenModel(db *gorm.DB, table []string, path string, dbDst *gorm.DB) {
	if len(table) == 0 {
		return
	}

	g := gen.NewGenerator(gen.Config{
		ModelPkgPath:      config.Root + path + "/model",
		OutPath:           config.Root + path,
		Mode:              gen.WithoutContext,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	})

	g.UseDB(db)

	var ok bool
	for _, t := range table {
		if strings.Contains(t, ".") {
			name := strings.Split(t, ".")[1]
			if dbDst != nil && dbDst.Migrator().HasTable(name) {
				continue
			}
			ok = true
			g.ApplyBasic(g.GenerateModelAs(t, tool.SepNameToCamel(name, true), gen.WithMethod((&defaultPgModel{}).TableName)))
		} else {
			ok = true
			g.ApplyBasic(g.GenerateModel(t))
		}
	}

	if ok {
		g.Execute()
	}
}
