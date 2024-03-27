package click

import (
	"fmt"
	"go-com/config"
	"go-com/core/orm"
	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

type Config struct {
	config.Clickhouse
}

func NewDb(cfg Config) *gorm.DB {
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s/%s?dial_timeout=10s&read_timeout=20s", cfg.User, cfg.Password, cfg.Addr, cfg.Dbname)
	return orm.NewDbSimple(clickhouse.Open(dsn))
}
