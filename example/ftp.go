package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/ftper"
	"go-com/core/logr"
	"os"
	"os/exec"
	"strings"
)

// 单文件下载示例
func singleFile() {
	var err error
	cfg := config.C.Ftp
	var path, filePrefix, filename string
	sh := fmt.Sprintf(ftper.SingleFileSh, strings.Split(cfg.Addr, ":")[0], cfg.User, cfg.Password, path, filename)
	fileShell := path + filePrefix + ".sh"
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
