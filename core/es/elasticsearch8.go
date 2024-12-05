package es

import (
	"bytes"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"go-com/core/logr"
	"net/http"
	"time"
)

type V8 struct {
	client *elasticsearch.Client
}

func NewEs8(cfg Config) V8 {
	var err error
	var v8 V8
	v8.client, err = elasticsearch.NewClient(elasticsearch.Config{
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
	return v8
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
	                            "lt": "2025-04-19T19:22:00",
								"time_zone": "+08:00"
	                        }
	                    }
	                }
	            ]
	        }
	    }
	}
*/
func (e V8) SearchPagination(index string, sql string, handle func(data []map[string]interface{})) int {
	es := e.client
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
Search 查询：相等 大小于 in 取反；聚合函数（"size":0）：avg max min sum

	{
	    "size": 10,
	    "aggs": {
	        "avg_of_delay": {
	            "avg": {
	                "field": "delay"
	            }
	        }
	    },
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
	            "minimum_should_match": 1,
	            "must_not": {
	                "range": {
	                    "age": {
	                        "gte": 10,
	                        "lte": 20
	                    }
	                }
	            },
	            "filter": [
	                {
	                    "term": {
	                        "eventData.host": "10.255.12.10"
	                    }
	                },
	                {
	                    "term": {
	                        "taskid": "942266"
	                    }
	                },
	                {
	                    "terms": {
	                        "timestamp": [
	                            "1719230478829",
	                            "1719230481822"
	                        ]
	                    }
	                },
	                {
	                    "range": {
	                        "delay": {
	                            "gt": 0
	                        }
	                    }
	                },
	                {
	                    "range": {
	                        "@timestamp": {
	                            "gte": "2024-06-24T09:22:00",
	                            "lt": "2025-06-25T19:22:00",
								"time_zone": "+08:00"
	                        }
	                    }
	                }
	            ]
	        }
	    },
		"sort":{
			"@timestamp":"desc"
		}
	}
*/
func (e V8) Search(index string, sql string) (int, []map[string]interface{}, map[string]map[string]interface{}) {
	es := e.client
	// 查询语句转buf
	buffer := ReqBufPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		ReqBufPool.Put(buffer)
	}()
	buffer.WriteString(sql)

	// 首次查询 search
	respSearch, err := es.Search(es.Search.WithIndex(index), es.Search.WithBody(buffer), es.Search.WithPretty())
	var base Base
	if err != nil {
		logr.L.Error(err)
		return 0, nil, nil
	}
	json.NewDecoder(respSearch.Body).Decode(&base)
	respSearch.Body.Close()
	return base.Hits.Total.Value, base.Hits.Hits, base.Aggregations
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
	                            "lt": "2024-05-20T14:58:55",
								"time_zone": "+08:00"
	                        }
	                    }
	                }
	            ]
	        }
	    }
	}
*/
func (e V8) Count(index string, sql string) int {
	es := e.client
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
	base := Base{}
	json.NewDecoder(respClient.Body).Decode(&base)
	respClient.Body.Close()
	return base.Count
}

func (e V8) CreateIndex(index string, sql string) {
	es := e.client
	es.Indices.Create("es-8-tt")
}
