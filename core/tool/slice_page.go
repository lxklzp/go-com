package tool

// SlicePage 切片分页
type SlicePage struct {
	totalCount int64
	page       int64
	pageSize   int64
	From       int64
	To         int64
}

func (sp *SlicePage) Before(totalCount int64, pageSize int64) {
	sp.page = 0
	sp.From = 0
	sp.To = 0
	sp.totalCount = totalCount
	sp.pageSize = pageSize
}

func (sp *SlicePage) Next() bool {
	if sp.To >= sp.totalCount {
		return false
	}
	sp.From = sp.page * sp.pageSize
	sp.page++
	if sp.totalCount <= sp.page*sp.pageSize {
		sp.To = sp.totalCount
	} else {
		sp.To = sp.page * sp.pageSize
	}
	return true
}
