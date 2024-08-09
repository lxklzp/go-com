package controller

import (
	"go-com/config"
)

type test struct {
}

func init() {
	config.AddRouterApi(test{}, &config.RouterApiWebList)
}
