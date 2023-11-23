package main

import (
	"fmt"
	"go-com/config"
	"go-com/global"
	"os"
	"os/exec"
	"path"
)

func main() {
	config.Load()
	global.InitLog("build")

	// 编译
	var err error
	root := config.Root
	buildPath := root + "runtime/build/"
	program := "main"
	cmd := exec.Command("sh", "-c", fmt.Sprintf("go build -o %s %s", root+program, root+"main.go"))
	if err = cmd.Run(); err != nil {
		global.Log.Fatal(err)
	}

	// 打包
	fileList := []string{
		program,
		"config/config.yaml",
	}
	if err = os.RemoveAll(buildPath); err != nil {
		global.Log.Fatal(err)
	}
	for _, file := range fileList {
		if err = os.MkdirAll(buildPath+path.Dir(file), 0777); err != nil {
			global.Log.Fatal(err)
		}
		global.CopyFile(buildPath+file, root+file)
	}
	if err = os.Remove(root + program); err != nil {
		global.Log.Error(err)
	}
}
