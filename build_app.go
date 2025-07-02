package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/filer"
	"go-com/core/logr"
	"os"
	"os/exec"
	"path"
)

func main() {
	config.Load()
	var err error
	root := config.Root

	/***** 配置区域 开始 *****/
	logr.InitLog("build_app")
	buildPath := root + "runtime/build_app/"
	program := "main_app"
	cmd := exec.Command("sh", "-c", fmt.Sprintf("go build -o %s %s", buildPath+program, root+program+".go"))
	fileList := []string{
		program,
		"config/config.yaml",
	}
	/***** 配置区域 结束 *****/

	if err = os.RemoveAll(buildPath); err != nil {
		logr.L.Fatal(err)
	}
	for _, file := range fileList {
		if err = os.MkdirAll(buildPath+path.Dir(file), 0755); err != nil {
			logr.L.Fatal(err)
		}
		filer.CopyFile(buildPath+file, root+file)
	}
	os.Mkdir(buildPath+"/runtime", 0777)
	if err = cmd.Run(); err != nil {
		logr.L.Fatal(err)
	}

	logr.L.Infof("打包成功：%s", buildPath+program)
}
