package main

import (
	"fmt"
	"go-com/config"
	"go-com/core/tool"
)

func main() {
	ll := []map[string]interface{}{
		{
			"kk":   "",
			"name": "aa0",
		},
		{
			"value": "",
			"name":  "aa1",
		},
		{
			"value": nil,
			"name":  "aa2",
		},
		{
			"value": "12.34",
			"name":  "aa3",
		},
		{
			"value": 500,
			"name":  "aa4",
		},
		{
			"value": 12.36,
			"name":  "aa5",
		},
		{
			"value": -100,
			"name":  "aa6",
		},
	}

	lll := tool.MapSort(ll, "value", tool.SortFieldTypeFloat, config.SortDesc)
	fmt.Println(lll)

	lls := []map[string]interface{}{
		{
			"value1": "aaa",
			"name":   "aa1",
		},
		{
			"value": nil,
			"name":  "aa2",
		},
		{
			"value": "abc",
			"name":  "aa3",
		},
		{
			"value": "bbc",
			"name":  "aa4",
		},
		{
			"value": "chsy1",
			"name":  "aa5",
		},
		{
			"value": "aaaaaaaa",
			"name":  "aa6",
		},
	}

	llls := tool.MapSort(lls, "value", tool.SortFieldTypeString, config.SortAsc)
	fmt.Println(llls)
}
