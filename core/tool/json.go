package tool

import (
	"bytes"
	"encoding/json"
	"go-com/config"
)

// SearchJsonByKeysRecursive 在json数据中递归查找指定键名相同的所有数据
func SearchJsonByKeysRecursive(object interface{}, key []string, handler func(object map[string]interface{}, key string)) {
	switch object.(type) {
	case []interface{}:
		object := object.([]interface{})
		for _, sub := range object {
			SearchJsonByKeysRecursive(sub, key, handler)
		}
	case map[string]interface{}:
		object := object.(map[string]interface{})
		for k, sub := range object {
			if SliceHas(key, k) {
				handler(object, k)
			} else {
				SearchJsonByKeysRecursive(sub, key, handler)
			}
		}
	}
}

// SearchJsonOnceByKey 在json数据中查找一次指定键名的值
func SearchJsonOnceByKey(object interface{}, key string) interface{} {
	if key == "" {
		return object
	}
	switch object.(type) {
	case []interface{}:
		object := object.([]interface{})
		for _, sub := range object {
			return SearchJsonOnceByKey(sub, key)
		}
	case map[string]interface{}:
		object := object.(map[string]interface{})
		for k, sub := range object {
			if k == key {
				return sub
			} else {
				if value := SearchJsonOnceByKey(sub, key); value != config.Sep {
					return value
				}
			}
		}
	}
	return config.Sep
}

// JsonEncode json编码，不转义字符
func JsonEncode(v interface{}) []byte {
	bsBuf := config.BufPool.Get().(*bytes.Buffer)
	defer func() {
		bsBuf.Reset()
		config.BufPool.Put(bsBuf)
	}()
	encoder := json.NewEncoder(bsBuf)
	encoder.SetEscapeHTML(false)
	encoder.Encode(v)

	src := bsBuf.Bytes()
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
