package tool

import (
	"fmt"
	"go-com/config"
	"reflect"
	"sort"
	"strings"
)

const (
	SortFieldTypeFloat  = 1 // float方式对数字排序
	SortFieldTypeString = 2 // string方式排序
)

// 根据float排序，升序
type sortFloat struct {
	value float64
	index int
}
type SortFloatList []sortFloat

// 根据string排序，升序
type sortString struct {
	value string
	index int
}
type SortStringList []sortString

// MapSort 对map数组按指定字段进行排序，sortType：1 升序 2 降序
func MapSort(mList []map[string]interface{}, sortField string, sortFieldType int, sortType int) []map[string]interface{} {
	// 参数验证
	if len(mList) == 0 || sortField == "" || (sortFieldType != SortFieldTypeFloat && sortFieldType != SortFieldTypeString) ||
		(sortType != config.SortAsc && sortType != config.SortDesc) {
		return nil
	}

	var ok bool
	result := make([]map[string]interface{}, len(mList))
	switch sortFieldType {
	case SortFieldTypeFloat:
		// 准备排序数据
		var err error
		var value, valueDefault float64
		if sortType == config.SortAsc {
			valueDefault = config.MaxFloat
		} else {
			valueDefault = config.MinFloat
		}
		// 构建排序结构体：SortFloatList
		sortFloatList := make(SortFloatList, len(mList))
		for k, m := range mList {
			if _, ok = m[sortField]; !ok {
				// 处理排序字段不存在的情况
				value = valueDefault
			} else if value, err = InterfaceToFloat(m[sortField]); err != nil {
				// 处理排序字段不能转换成float64的情况
				value = valueDefault
			}
			sortFloatList[k] = sortFloat{value: value, index: k}
		}

		if sortType == config.SortAsc {
			// 升序排序
			sort.Slice(sortFloatList, func(i, j int) bool {
				return sortFloatList[i].value < sortFloatList[j].value
			})
		} else {
			// 降序排序
			sort.Slice(sortFloatList, func(i, j int) bool {
				return sortFloatList[i].value > sortFloatList[j].value
			})
		}

		// 根据已排序的SortFloatList，生成结果数据
		for k, si := range sortFloatList {
			result[k] = mList[si.index]
		}
	case SortFieldTypeString:
		// 准备排序数据
		var value, valueDefault string
		if sortType == config.SortAsc {
			valueDefault = string([]byte{127, 127, 127, 127, 127, 127, 127, 127})
		} else {
			valueDefault = ""
		}
		// 构建排序结构体：SortStringList
		var sortStringList SortStringList
		sortStringList = make([]sortString, len(mList))
		for k, m := range mList {
			if _, ok = m[sortField]; !ok {
				// 处理排序字段不存在的情况
				value = valueDefault
			} else {
				value = InterfaceToString(m[sortField])
			}
			sortStringList[k] = sortString{value: value, index: k}
		}

		// SortStringList排序
		if sortType == config.SortAsc {
			sort.Slice(sortStringList, func(i, j int) bool {
				return strings.Compare(sortStringList[i].value, sortStringList[j].value) < 0
			})
		} else {
			sort.Slice(sortStringList, func(i, j int) bool {
				return strings.Compare(sortStringList[i].value, sortStringList[j].value) > 0
			})
		}

		// 根据已排序的SortStringList，生成结果数据
		for k, si := range sortStringList {
			result[k] = mList[si.index]
		}
	}

	return result
}

// MapIntersect 求map[T]bool的交集
func MapIntersect[T int | string](a map[T]bool, b map[T]bool) map[T]bool {
	res := make(map[T]bool)
	for v := range a {
		if _, ok := b[v]; ok {
			res[v] = true
		}
	}
	return res
}

// StructToMap 结构体转map
func StructToMap(in interface{}, tagName string) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	v := reflect.ValueOf(in)
	if v.Kind() != reflect.Struct { // 非结构体返回错误提示
		return nil, fmt.Errorf("StructToMap 只支持结构体，不支持 %T", v)
	}

	t := v.Type()
	// 遍历结构体字段
	// 指定tagName值为map中key;字段值为map中value
	for i := 0; i < v.NumField(); i++ {
		fi := t.Field(i)
		// 递归处理嵌套结构体
		typeName := t.Field(i).Type.Name()
		if v.Field(i).Kind() == reflect.Struct && typeName != "Timestamp" && typeName != "JSON" {
			outChild, _ := StructToMap(v.Field(i).Interface(), tagName)
			for childK, childV := range outChild {
				out[childK] = childV
			}
		} else if tagValue := fi.Tag.Get(tagName); tagValue != "" {
			out[tagValue] = v.Field(i).Interface()
		}
	}
	return out, nil
}
