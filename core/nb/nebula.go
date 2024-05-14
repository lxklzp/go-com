package nb

import (
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"go-com/config"
	"go-com/core/logr"
)

type Config struct {
	config.Nebula
}

func NewNebula(cfg Config) *nebula.SessionPool {
	conf, err := nebula.NewSessionPoolConf(
		cfg.User,
		cfg.Password,
		[]nebula.HostAddress{{Host: cfg.Host, Port: cfg.Port}},
		cfg.Dbname,
	)
	if err != nil {
		logr.L.Fatal(err)
	}

	db, err := nebula.NewSessionPool(*conf, nebula.DefaultLogger{})
	if err != nil {
		logr.L.Fatal(err)
	}
	return db
}
