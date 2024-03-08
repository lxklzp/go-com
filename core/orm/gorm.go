package orm

import (
	"fmt"
	"go-com/config"
	"go-com/core/logr"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

type DbConfig struct {
	config.DbConfig
}

func NewDb(dialector gorm.Dialector, cfg DbConfig) *gorm.DB {
	var err error
	var db *gorm.DB

	// 配置gorm的日志
	var logLevel logger.LogLevel
	if config.C.App.Environment == config.EnvDev {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}
	newLogger := logger.New(
		log.New(logr.L.Out, fmt.Sprintf("[db] "), log.Lmsgprefix),
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 连接数据库
	db, err = gorm.Open(dialector, &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			TablePrefix:   cfg.Prefix,
		},
		Logger:                                   newLogger,
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		logr.L.Fatal(err)
	}

	// 连接配置
	sqlDB, err := db.DB()
	if err != nil {
		logr.L.Fatal(err)
	}
	// 设置最大连接数
	if cfg.MaxOpenConns != 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	// 设置最大空闲连接数
	if cfg.MaxIdleConns != 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	// 设置一个连接的最大存活时长
	if cfg.ConnMaxLifetime != 0 {
		sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	}
	return db
}

func NewDbSimple(dialector gorm.Dialector) *gorm.DB {
	var err error
	var db *gorm.DB

	// 配置gorm的日志
	var logLevel logger.LogLevel
	if config.C.App.Environment == config.EnvDev {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}
	newLogger := logger.New(
		log.New(logr.L.Out, fmt.Sprintf("[db] "), log.Lmsgprefix),
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// 连接数据库
	db, err = gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		logr.L.Fatal(err)
	}
	return db
}
