package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/pkg/errors"
	"go-com/config"
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
	respSearch, err := es.Search(es.Search.WithIndex(index), es.Search.WithScroll(time.Minute*3), es.Search.WithBody(buffer), es.Search.WithPretty())
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
	                            "lt": "2025-06-25T19:22:00"
	                        }
	                    }
	                }
	            ]
	        }
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

	// 首次查询 search
	respSearch, err := es.Search(es.Search.WithIndex(index), es.Search.WithBody(buffer), es.Search.WithPretty())
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
	respClient, err := es.Count(es.Count.WithIndex(index), es.Count.WithBody(buffer))
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
	resp, err := es.Indices.Create(index, es.Indices.Create.WithBody(strings.NewReader(body)))
	logr.L.Debug(resp)
	return err
}

func DeleteIndex(es *elasticsearch.Client, index string) error {
	resp, err := es.Indices.Delete([]string{index})
	logr.L.Debug(resp)
	return err
}

func IndexExists(es *elasticsearch.Client, index string) bool {
	resp, err := es.Indices.Exists([]string{index})

	if err != nil {
		logr.L.Error(err)
		return false
	}
	if resp == nil {
		logr.L.Error("elasticsearch indices-exists 的返回结果为空")
		return false
	}

	if resp.StatusCode == 200 {
		return true
	} else {
		return false
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
	bsBuf := config.BufPool.Get().(*bytes.Buffer)
	defer func() {
		bsBuf.Reset()
		config.BufPool.Put(bsBuf)
	}()
	encoder := json.NewEncoder(bsBuf)
	encoder.SetEscapeHTML(false)

	var dataJson []byte
	var i int
	pageSize := 1000
	timestamp := time.Now().Format(config.DateTimeStandardFormatter)
	for _, d := range data {
		i++
		d["@timestamp"] = timestamp
		encoder.Encode(d)
		dataJson = append(dataJson, []byte(fmt.Sprintf(`{"index":{"_id":"%d"}}%s`, tool.SnowflakeComm.GetId(), "\n"))...)
		dataJson = append(dataJson, bsBuf.Bytes()...)
		bsBuf.Reset()
		dataJson = append(dataJson, "\n"...)
		if i >= pageSize {
			insert(es, index, dataJson)
			dataJson = dataJson[:0]
			i = 0
			timestamp = time.Now().Format(config.DateTimeStandardFormatter)
		}
	}

	if len(dataJson) > 0 {
		insert(es, index, dataJson)
	}
}

// 批量写入一轮
func insert(es *elasticsearch.Client, index string, dataJson []byte) {
	res, err := es.Bulk(bytes.NewReader(dataJson), es.Bulk.WithIndex(index))
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
	res, err := es.DeleteByQuery([]string{index}, bytes.NewReader([]byte(sql)))
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
