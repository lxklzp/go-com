package global

import (
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"go-com/config"
)

var Nebula *nebula.SessionPool

func InitNebula() {
	cfg := config.C.Nebula
	conf, err := nebula.NewSessionPoolConf(
		cfg.User,
		cfg.Password,
		[]nebula.HostAddress{{Host: cfg.Host, Port: cfg.Port}},
		cfg.Dbname,
	)
	if err != nil {
		Log.Fatal(err)
	}

	if Nebula, err = nebula.NewSessionPool(*conf, nebula.DefaultLogger{}); err != nil {
		Log.Fatal(err)
	}
}
