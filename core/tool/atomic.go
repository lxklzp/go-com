package tool

import (
	"sync/atomic"
)

// AtomicIncr 原子性的自增，带阈值
func AtomicIncr(count *atomic.Int64, step int64, countMax int64) bool {
	for {
		countNow := count.Load()
		if countNow >= countMax {
			return false
		}
		if count.CompareAndSwap(countNow, countNow+step) {
			return true
		}
	}
}

// AtomicDecr 原子性的自减，带阈值
func AtomicDecr(count *atomic.Int64, step int64, countMin int64) bool {
	for {
		countNow := count.Load()
		if countNow <= countMin {
			return false
		}
		if count.CompareAndSwap(countNow, countNow-step) {
			return true
		}
	}
}
