package excel

import (
	"github.com/xuri/excelize/v2"
	"go-com/core/logr"
)

// ReadStandard 标准读取excel
func ReadStandard(filename, sheet string, from int) [][]string {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		logr.L.Fatal(err)
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		logr.L.Fatal(err)
	}

	return rows[from:]
}
