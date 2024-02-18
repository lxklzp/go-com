package global

import (
	"fmt"
	"go-com/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

var GormPg *gorm.DB

func InitGormPg() {
	var err error

	// sql日志
	var logLevel logger.LogLevel
	if config.C.App.Environment == config.EnvDev {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	newLogger := logger.New(
		log.New(Log.Out, fmt.Sprintf("[pgsql] "), log.Lmsgprefix),
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	cfg := config.C.Pgsql
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", cfg.Host, cfg.User, cfg.Password, cfg.Dbname, cfg.Port)
	GormPg, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		Log.Fatal(err)
	}

	sqlDB, err := GormPg.DB()
	if err != nil {
		Log.Fatal(err)
	}
	// 用于设置连接池中空闲连接的最大数量。
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	// 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
}
