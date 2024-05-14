package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/nb"
	"go-com/internal/app"
)

func main() {
	config.Load()
	logr.InitLog("test")
	app.Nb = nb.NewNebula(nb.Config{Nebula: config.C.Nebula})
	res, err := app.Nb.Execute("MATCH ()-[e]->() RETURN e limit 100;")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(res.GetRows()[0])

}
