package rpc

import (
	"bytes"
	"sync"
)

var reqBufPool *sync.Pool

func init() {
	reqBufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4096))
		},
	}
}
