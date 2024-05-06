package main

import (
	"context"
	"fmt"
	"github.com/expr-lang/expr"
	"go-com/config"
	"go-com/core/logr"
	"go-com/core/pg"
	"go-com/internal/app"
)

func main() {
	config.Load()
	logr.InitLog("web")
	app.Pg = pg.NewDb(pg.Config{Postgresql: config.C.Postgresql})

	c := 2.561472e+06
	fmt.Printf("%f\n", c)

	// 表达式引擎示例
	exprCode := `let v = 2561473;
v >= 3.561472e+06 ? 3 : (v > 2.561472e+06 ? 2 : (v == 3.561472e+06 ? 1 : 0))`
	logr.L.Debug(exprCode)
	program, err := expr.Compile(exprCode)
	if err != nil {
		logr.L.Error(err)
		return
	}
	output, err := expr.Run(program, nil)
	if err != nil {
		logr.L.Error(err)
		return
	}
	fmt.Println(output)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	<-ctx.Done()
}
