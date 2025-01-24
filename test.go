package main

import (
	"fmt"
	"sync/atomic"
)

func main() {
	supfileindex := int32(2)
	atomic.CompareAndSwapInt32(&supfileindex, int32(2), 1)
	fmt.Println(supfileindex)
}
