package global

import (
	"database/sql"
	"fmt"
	"go-com/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

var GormMy *gorm.DB

func InitGormMy() {
	cfg := config.C.Mysql
	Db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Addr, cfg.Dbname))
	if err != nil {
		Log.Fatal(err)
	}

	Db.SetConnMaxLifetime(time.Second * cfg.ConnMaxLifetime)
	Db.SetMaxOpenConns(cfg.MaxOpenConns)
	Db.SetMaxIdleConns(cfg.MaxIdleConns)

	// sql日志
	var logLevel logger.LogLevel
	if config.C.App.Environment == config.EnvDev {
		logLevel = logger.Info
	} else {
		logLevel = logger.Error
	}

	newLogger := logger.New(
		log.New(Log.Out, fmt.Sprintf("[mysql] "), log.Lmsgprefix),
		logger.Config{
			SlowThreshold:             time.Second, // 慢SQL阈值
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	GormMy, err = gorm.Open(mysql.New(mysql.Config{
		Conn: Db,
	}), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		Log.Fatal(err)
	}
}
