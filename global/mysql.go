package global

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"go-com/config"
	"time"
)

var Db *sql.DB

func InitDb() {
	var err error
	cfg := config.C.Mysql
	Db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", cfg.User, cfg.Password, cfg.Addr, cfg.Db))
	if err != nil {
		Log.Fatal(err)
	}

	Db.SetConnMaxLifetime(time.Second * cfg.ConnMaxLifetime)
	Db.SetMaxOpenConns(cfg.MaxOpenConns)
	Db.SetMaxIdleConns(cfg.MaxIdleConns)
}

func DbTransaction() *sql.Tx {
	tx, err := Db.Begin()
	if err != nil {
		if tx != nil {
			tx.Rollback()
		}
		Log.Panic(err)
	}
	return tx
}
