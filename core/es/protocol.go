package es

import (
	"bytes"
	"go-com/config"
	"sync"
)

var ReqBufPool *sync.Pool

type Config struct {
	config.Es
}

func init() {
	ReqBufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 4096))
		},
	}
}

type Base struct {
	ScrollId     string `json:"_scroll_id"`
	Hits         Hits   `json:"hits"`
	Aggregations map[string]map[string]interface{}
	Count        int `json:"count"`
}

type Hits struct {
	Total HitsTotal                `json:"total"`
	Hits  []map[string]interface{} `json:"hits"`
}

type HitsTotal struct {
	Value int `json:"value"`
}

type Aggregations struct {
}
