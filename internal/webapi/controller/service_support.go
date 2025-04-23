package controller

import (
	"go-com/config"
)

/********** 服务支持 **********/

type ss struct {
}

func init() {
	config.AddRouterApi(ss{}, &config.RouterWebApiList)
}
