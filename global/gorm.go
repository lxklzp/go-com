package global

import (
	"fmt"
	"go-com/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"log"
	"time"
)

var Gorm *gorm.DB

func InitGorm() {
	var err error

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

	Gorm, err = gorm.Open(mysql.New(mysql.Config{
		Conn: Db,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		Log.Fatal(err)
	}
}
