package dp

import "fmt"

// 装饰器 这是一个后进先执行的方法栈，形成一条处理链。

func main() {
	c := chain{wrapperList: []wrapper{a(), b(), c()}}
	c.run(func() {
		fmt.Println("hello world")
	})
}

type run func()

type wrapper func(run) run

type chain struct {
	wrapperList []wrapper
}

func (c chain) run(r run) run {
	for i := range c.wrapperList {
		r = c.wrapperList[len(c.wrapperList)-i-1](r)
	}
	return r
}

func a() wrapper {
	return func(r run) run {
		fmt.Println("a")
		return r
	}
}

func b() wrapper {
	return func(r run) run {
		fmt.Println("b")
		return r
	}
}

func c() wrapper {
	return func(r run) run {
		fmt.Println("c")
		return r
	}
}
