package tool

import "strconv"

// SliceHas 值在切片中是否存在
func SliceHas[T int | string](array []T, value T) bool {
	if len(array) == 0 {
		return false
	}

	for _, item := range array {
		if item == value {
			return true
		}
	}
	return false
}

// SliceUnique 切片去重
func SliceUnique[T int | string](array []T) []T {
	if len(array) == 0 {
		return nil
	}

	m := make(map[T]struct{})
	var ok bool
	var result []T
	for _, v := range array {
		if _, ok = m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		result = append(result, v)
	}
	return result
}

// SliceRemoveValue 按值删除切片元素
func SliceRemoveValue[T int | float64 | string](array []T, val T) []T {
	if len(array) == 0 {
		return nil
	}

	var result []T
	for _, v := range array {
		if v != val {
			result = append(result, v)
		}
	}
	return result
}

// SliceAvg 求平均值
func SliceAvg[T int | float64](array []T) T {
	if len(array) == 0 {
		return T(0)
	}

	var sum T
	for _, v := range array {
		sum += v
	}
	return sum / T(len(array))
}

// SliceIntToString 数字切片转字符串切片
func SliceIntToString(sliceInt []int) []string {
	if len(sliceInt) == 0 {
		return nil
	}

	var sliceString []string
	for _, i := range sliceInt {
		sliceString = append(sliceString, strconv.Itoa(i))
	}
	return sliceString
}

// SliceStringToInt 字符串切片转数字切片
func SliceStringToInt(sliceString []string) []int {
	if len(sliceString) == 0 {
		return nil
	}

	var sliceInt []int
	for _, s := range sliceString {
		i, _ := strconv.Atoi(s)
		sliceInt = append(sliceInt, i)
	}
	return sliceInt
}

// SliceIsSubset 判断是否是子集
func SliceIsSubset[T int | string](sub []T, super []T) bool {
	superMap := make(map[T]bool)
	for _, v := range super {
		superMap[v] = true
	}
	for _, v := range sub {
		if !superMap[v] {
			return false
		}
	}
	return true
}

// SliceEqualUnordered 判断两个切片是否包含完全相同元素（不考虑顺序）
func SliceEqualUnordered[T int | string](a []T, b []T) bool {
	if len(a) != len(b) {
		return false
	}

	// 将a切片值录入计数map
	valueCount := make(map[T]int, len(a))
	for _, v := range a {
		valueCount[v]++
	}

	// 比对b切片值计数
	for _, v := range b {
		// a中没有b中的元素
		if valueCount[v] == 0 {
			return false
		}
		// a中有b中的元素，兑掉一次
		valueCount[v]--
		// a中兑完则删除
		if valueCount[v] == 0 {
			delete(valueCount, v)
		}
	}

	// b中没有a中的元素
	if len(valueCount) != 0 {
		return false
	}

	return true
}
