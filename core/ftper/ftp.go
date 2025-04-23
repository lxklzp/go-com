package ftper

import (
	"fmt"
	"github.com/jlaffaye/ftp"
	"go-com/config"
	"go-com/core/logr"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	SingleFileSh = `#!/bin/sh
ftp -v -n %s<<EOF
user %s %s
binary
lcd %s
prompt
get "%s"
bye
EOF
echo "download from ftp successfully"` // shell ftp 文件下载脚本
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

// DownloadSingleFileByShell 通过 shell ftp 下载单个文件
func DownloadSingleFileByShell(dstPath string, filePrefix string, filename string) {
	var err error
	if err = os.MkdirAll(dstPath, 0755); err != nil {
		logr.L.Error(err)
		return
	}
	cfg := config.C.Ftp
	sh := fmt.Sprintf(SingleFileSh, strings.Split(cfg.Addr, ":")[0], cfg.User, cfg.Password, dstPath, filename)
	fileShell := dstPath + filePrefix + ".sh"
	if err = os.WriteFile(fileShell, []byte(sh), 0777); err != nil {
		logr.L.Error(err)
		return
	}
	// 执行 shell ftp 文件下载脚本
	cmd := exec.Command(fileShell)
	res, err := cmd.Output()
	if err != nil {
		logr.L.Error(err)
		return
	}
	logr.L.Debug(string(res))
}
