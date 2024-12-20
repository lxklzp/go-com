package ftper

import (
	"github.com/jlaffaye/ftp"
	"go-com/config"
	"go-com/core/logr"
	"time"
)

const (
	SingleFileSh = `
#!/bin/sh
ftp -v -n %s<<EOF
user %s %s
binary
lcd %s
prompt
get "%s"
bye
EOF
echo "download from ftp successfully"
`
)

type Config struct {
	config.Ftp
}

func NewFtp(cfg Config) *ftp.ServerConn {
	cli, err := ftp.Dial(cfg.Addr, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		logr.L.Fatal(err)
	}
	err = cli.Login(cfg.User, cfg.Password)
	if err != nil {
		logr.L.Fatal(err)
	}
	return cli
}
