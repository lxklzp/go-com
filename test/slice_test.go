package test

import (
	"go-com/core/tool"
	"testing"
)

func TestSliceEqualUnordered(t *testing.T) {
	tests := []struct {
		name string
		a    []int
		b    []int
		want bool
	}{
		// 基本测试用例
		{
			name: "相同顺序的切片",
			a:    []int{1, 2, 3},
			b:    []int{1, 2, 3},
			want: true,
		},
		{
			name: "不同顺序的相同切片",
			a:    []int{1, 2, 3},
			b:    []int{3, 2, 1},
			want: true,
		},
		{
			name: "不同元素",
			a:    []int{1, 2, 3},
			b:    []int{1, 2, 4},
			want: false,
		},
		{
			name: "长度不同",
			a:    []int{1, 2, 3},
			b:    []int{1, 2},
			want: false,
		},

		// 重复元素测试
		{
			name: "有重复元素的相同切片",
			a:    []int{1, 2, 2, 3},
			b:    []int{2, 1, 2, 3},
			want: true,
		},
		{
			name: "重复元素数量不同",
			a:    []int{1, 2, 2, 3},
			b:    []int{1, 2, 3, 3},
			want: false,
		},

		// 边界测试
		{
			name: "空切片",
			a:    []int{},
			b:    []int{},
			want: true,
		},
		{
			name: "nil切片",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "一个nil一个空",
			a:    nil,
			b:    []int{},
			want: true,
		},
		{
			name: "一个nil一个有元素",
			a:    nil,
			b:    []int{1},
			want: false,
		},

		// 大切片测试
		{
			name: "大切片相同",
			a:    makeRange(1, 1000),
			b:    makeRange(1, 1000),
			want: true,
		},
		{
			name: "大切片不同",
			a:    makeRange(1, 1000),
			b:    append(makeRange(1, 999), 1001),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tool.SliceEqualUnordered(tt.a, tt.b); got != tt.want {
				t.Errorf("SliceEqualUnordered() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 测试字符串类型的切片
func TestSliceEqualUnorderedString(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "字符串相同",
			a:    []string{"a", "b", "c"},
			b:    []string{"c", "b", "a"},
			want: true,
		},
		{
			name: "字符串不同",
			a:    []string{"a", "b", "c"},
			b:    []string{"a", "b", "d"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tool.SliceEqualUnordered(tt.a, tt.b); got != tt.want {
				t.Errorf("SliceEqualUnordered() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 辅助函数：生成一个整数范围切片
func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}
