package main

import (
	"fmt"
	"github.com/expr-lang/expr"
	"go-com/core/logr"
)

func Expr() {
	// 表达式引擎示例
	exprCode := `let v = 10;
v > 30 ? 3 : (v > 20 ? 2 : (v > 10 ? 1 : 0))`
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
}
