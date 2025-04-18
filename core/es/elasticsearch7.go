package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/pkg/errors"
	"go-com/core/logr"
	"go-com/core/tool"
	"net/http"
	"strings"
	"sync"
	"time"
)

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
	logr.L.Debug(sql)
	// 查询语句转buf
	buffer := ReqBufPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		ReqBufPool.Put(buffer)
	}()
	buffer.WriteString(sql)

	// 首次查询 search
	ctx, ctxCancel := context.WithTimeout(context.TODO(), time.Second*30)
	respSearch, err := es.Search(es.Search.WithIndex(index), es.Search.WithScroll(time.Minute*3), es.Search.WithBody(buffer), es.Search.WithPretty(), es.Search.WithContext(ctx))
	ctxCancel()
	var base Base
	if err != nil {
		logr.L.Error(err)
		return 0
	}
	json.NewDecoder(respSearch.Body).Decode(&base)
	respSearch.Body.Close()
	total := base.Hits.Total.Value
	logr.L.Debug(total)
	handle(base.Hits.Hits)

	// 分页查询 scroll
	var respClient *esapi.Response
	for len(base.Hits.Hits) != 0 {
		ctx, ctxCancel = context.WithTimeout(context.TODO(), time.Second*30)
		respClient, err = es.Scroll(es.Scroll.WithScrollID(base.ScrollId), es.Scroll.WithScroll(time.Minute*3), es.Scroll.WithContext(ctx))
		ctxCancel()
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
	ctx, ctxCancel = context.WithTimeout(context.TODO(), time.Second*5)
	respClient, err = es.ClearScroll(es.ClearScroll.WithScrollID(base.ScrollId), es.ClearScroll.WithContext(ctx))
	ctxCancel()
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
	                            "lt": "2025-06-25T19:22:00"
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
func Search(es *elasticsearch.Client, index string, sql string) (int, []map[string]interface{}, map[string]map[string]interface{}) {
	logr.L.Debug(sql)
	// 查询语句转buf
	buffer := ReqBufPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		ReqBufPool.Put(buffer)
	}()
	buffer.WriteString(sql)

	ctx, ctxCancel := context.WithTimeout(context.TODO(), time.Second*30)
	respSearch, err := es.Search(es.Search.WithIndex(index), es.Search.WithBody(buffer), es.Search.WithPretty(), es.Search.WithContext(ctx))
	ctxCancel()
	var base Base
	if err != nil {
		logr.L.Error(err)
		return 0, nil, nil
	}
	json.NewDecoder(respSearch.Body).Decode(&base)
	respSearch.Body.Close()
	logr.L.Debug(base.Hits.Total.Value)
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
								"time_zone": "+08:00",
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
	logr.L.Debug(sql)
	// 查询语句转buf
	buffer := ReqBufPool.Get().(*bytes.Buffer)
	defer func() {
		buffer.Reset()
		ReqBufPool.Put(buffer)
	}()
	buffer.WriteString(sql)
	ctx, ctxCancel := context.WithTimeout(context.TODO(), time.Second*5)
	respClient, err := es.Count(es.Count.WithIndex(index), es.Count.WithBody(buffer), es.Count.WithContext(ctx))
	ctxCancel()
	if err != nil {
		logr.L.Error(err)
		return 0
	}
	base := Base{}
	json.NewDecoder(respClient.Body).Decode(&base)
	respClient.Body.Close()
	logr.L.Debug(base)
	return base.Count
}

/*
CreateIndex 创建索引

	{
	    "mappings": {
	        "properties": {
	            "@timestamp": {
	                "type": "date"
	            }
	        }
	    }
	}
*/
func CreateIndex(es *elasticsearch.Client, index string, body string) error {
	ctx, ctxCancel := context.WithTimeout(context.TODO(), time.Second*5)
	_, err := es.Indices.Create(index, es.Indices.Create.WithBody(strings.NewReader(body)), es.Indices.Create.WithContext(ctx))
	ctxCancel()
	return err
}

func DeleteIndex(es *elasticsearch.Client, index string) error {
	ctx, ctxCancel := context.WithTimeout(context.TODO(), time.Second*5)
	_, err := es.Indices.Delete([]string{index}, es.Indices.Delete.WithContext(ctx))
	ctxCancel()
	return err
}

func ExistsIndex(es *elasticsearch.Client, index string) (bool, error) {
	ctx, ctxCancel := context.WithTimeout(context.TODO(), time.Second*5)
	respClient, err := es.Indices.Exists([]string{index}, es.Indices.Exists.WithContext(ctx))
	ctxCancel()
	if err != nil {
		return false, err
	}
	if respClient.StatusCode == 200 {
		return true, nil
	} else {
		return false, nil
	}
}

type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"reason"`
				} `json:"caused_by"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}

// BatchInsert 批量写入
func BatchInsert(es *elasticsearch.Client, index string, data []map[string]interface{}) {
	var dataJson []byte
	var i int
	pageSize := 5000
	for _, d := range data {
		i++
		dataJson = append(dataJson, []byte(fmt.Sprintf(`{"index":{"_id":"%d"}}%s`, tool.SnowflakeComm.GetId(), "\n"))...)
		dataJson = append(dataJson, tool.JsonEncode(d)...)
		dataJson = append(dataJson, "\n"...)
		if i >= pageSize {
			insert(es, index, dataJson)
			dataJson = dataJson[:0]
			i = 0
		}
	}

	if len(dataJson) > 0 {
		insert(es, index, dataJson)
	}
}

// 批量写入一轮
func insert(es *elasticsearch.Client, index string, dataJson []byte) {
	ctx, ctxCancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer ctxCancel()
	res, err := es.Bulk(bytes.NewReader(dataJson), es.Bulk.WithIndex(index), es.Bulk.WithContext(ctx))
	var blk *bulkResponse
	var raw map[string]interface{}
	if err != nil {
		logr.L.Error(err)
		return
	}
	if res == nil {
		logr.L.Error("elasticsearch bulk 写入的返回结果为空")
		return
	}
	if res.IsError() {
		if err = json.NewDecoder(res.Body).Decode(&raw); err != nil {
			logr.L.Error(err)
			return
		} else {
			logr.L.Errorf("[%d] %s: %s", res.StatusCode,
				raw["error"].(map[string]interface{})["type"],
				raw["error"].(map[string]interface{})["reason"])
			return
		}
	} else {
		if err = json.NewDecoder(res.Body).Decode(&blk); err != nil {
			logr.L.Error(err)
			return
		} else {
			for _, d := range blk.Items {
				if d.Index.Status > 201 {
					logr.L.Errorf("[%d]: %s: %s: %s: %s", d.Index.Status,
						d.Index.Error.Type,
						d.Index.Error.Reason,
						d.Index.Error.Cause.Type,
						d.Index.Error.Cause.Reason)
				}
			}
		}
	}

	res.Body.Close()
}

func Delete(es *elasticsearch.Client, index string, sql string) error {
	ctx, ctxCancel := context.WithTimeout(context.TODO(), time.Second*10)
	res, err := es.DeleteByQuery([]string{index}, bytes.NewReader([]byte(sql)), es.DeleteByQuery.WithContext(ctx))
	ctxCancel()
	if err != nil {
		return err
	}
	if res == nil {
		return errors.New("elasticsearch delete_by_query 返回结果为空")
	}
	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("%s", res))
	}
	return nil
}
