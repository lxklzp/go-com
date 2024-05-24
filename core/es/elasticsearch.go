package es

import (
	"bytes"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"go-com/config"
	"go-com/core/logr"
	"net/http"
	"sync"
	"time"
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

func NewEs(cfg Config) *elasticsearch.Client {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.Addr,
		Username:  cfg.User,
		Password:  cfg.Password,
		Transport: &http.Transport{
			MaxIdleConnsPerHost: cfg.DbConfig.MaxIdleConns,
			MaxConnsPerHost:     cfg.DbConfig.MaxOpenConns,
		},
	})
	if err != nil {
		logr.L.Fatal(err)
	}
	return es
}

type Base struct {
	ScrollId string `json:"_scroll_id"`
	Hits     Hits   `json:"hits"`
	Count    int    `json:"count"`
}

type Hits struct {
	Total HitsTotal                `json:"total"`
	Hits  []map[string]interface{} `json:"hits"`
}

type HitsTotal struct {
	Value int `json:"value"`
}

/*
SearchPagination 查询分页函数 search + scroll
sql示例：

	{
		"size": 3000,
	    "query": {
	        "bool": {
				"should": [
	                {
	                    "range": {
	                        "status": {
	                            "gte": "400",
	                            "lte": "999"
	                        }
	                    }
	                },
	                {
	                    "term": {
	                        "status": "失败"
	                    }
	                }
	            ],
	            "minimum_should_match" : 1,
	            "filter": [
	                {
	                    "term": {
	                        "eventData.host": ""
	                    }
	                },
	                {
	                    "range": {
	                        "@timestamp": {
	                            "gte": "2024-04-18T19:22:00",
	                            "lt": "2025-04-19T19:22:00"
	                        }
	                    }
	                }
	            ]
	        }
	    }
	}
*/
func SearchPagination(es *elasticsearch.Client, index string, sql string, handle func(data []map[string]interface{})) int {
	// 查询语句转buf
	buffer := ReqBufPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		ReqBufPool.Put(buffer)
	}()
	buffer.WriteString(sql)

	// 首次查询 search
	respSearch, err := es.Search(es.Search.WithIndex(index), es.Search.WithScroll(time.Minute*3), es.Search.WithBody(buffer), es.Search.WithPretty())
	var base Base
	if err != nil {
		logr.L.Error(err)
		return 0
	}
	json.NewDecoder(respSearch.Body).Decode(&base)
	respSearch.Body.Close()
	total := base.Hits.Total.Value
	handle(base.Hits.Hits)

	// 分页查询 scroll
	var respClient *esapi.Response
	for len(base.Hits.Hits) != 0 {
		respClient, err = es.Scroll(es.Scroll.WithScrollID(base.ScrollId), es.Scroll.WithScroll(time.Minute*3))
		if err != nil {
			logr.L.Error(err)
			continue
		}
		base = Base{}
		json.NewDecoder(respClient.Body).Decode(&base)
		respClient.Body.Close()
		handle(base.Hits.Hits)
	}

	// 关闭 scroll
	respClient, err = es.ClearScroll(es.ClearScroll.WithScrollID(base.ScrollId))
	if err != nil {
		logr.L.Error(err)
	}
	return total
}

/*
Count 查询总条数 count

sql示例：

	{
	    "query": {
	        "bool": {
	            "filter": [
	                {
	                    "term": {
	                        "eventData.host": "10.255.248.141"
	                    }
	                },
	                {
	                    "range": {
	                        "@timestamp": {
	                            "gte": "2024-05-19T14:58:55",
	                            "lt": "2024-05-20T14:58:55"
	                        }
	                    }
	                }
	            ]
	        }
	    }
	}
*/
func Count(es *elasticsearch.Client, index string, sql string) int {
	// 查询语句转buf
	buffer := ReqBufPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		ReqBufPool.Put(buffer)
	}()
	buffer.WriteString(sql)
	respClient, err := es.Count(es.Count.WithIndex(index), es.Count.WithBody(buffer))
	if err != nil {
		logr.L.Error(err)
		return 0
	}
	logr.L.Info(respClient)
	base := Base{}
	json.NewDecoder(respClient.Body).Decode(&base)
	respClient.Body.Close()
	return base.Count
}
