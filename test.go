package main

import (
	"fmt"
	"go-com/core/tool"
)

func main() {

	a := []int{10, 30, 20, 40}
	b := []int{20, 40, 1, 10, 30, 10}

	fmt.Println(tool.SliceEqualUnordered(a, b))
}
