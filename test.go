package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/logr"
	"os"
	"os/exec"
)

func main() {
	config.Load()
	logr.InitLog("test")

	var err error
	cmd := exec.Command("sh", "-c", "cal")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		fmt.Println(err)
	}

}
