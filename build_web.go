package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/logr"
	"os/exec"
)

func main() {
	config.Load()
	var err error
	root := config.Root

	/***** 配置区域 开始 *****/
	logr.InitLog("build_web")
	buildPath := root + "runtime/build_app/"
	program := "main_web"
	cmd := exec.Command("sh", "-c", fmt.Sprintf("go build -o %s %s", buildPath+program, root+program+".go"))
	/***** 配置区域 结束 *****/

	if err = cmd.Run(); err != nil {
		logr.L.Fatal(err)
	}
	logr.L.Infof("打包成功：%s", buildPath+program)
}
