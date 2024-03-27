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
	logr.InitLog("build_web")
	reload := true // 是否完全重新打包
	buildPath := root + "runtime/build_web/"
	program := "main_web"
	cmd := exec.Command("sh", "-c", fmt.Sprintf("go build -o %s %s", root+program, root+"main_web.go"))
	fileList := []string{
		program,
		"config/config.yaml",
	}
	/***** 配置区域 结束 *****/

	if err = cmd.Run(); err != nil {
		logr.L.Fatal(err)
	}
	if reload {
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
	} else {
		os.Remove(buildPath + program)
		filer.CopyFile(buildPath+program, root+program)
	}
	if err = os.Remove(root + program); err != nil {
		logr.L.Error(err)
	}

	logr.L.Infof("打包成功：%s", buildPath+program)
}
