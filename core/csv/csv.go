package csv

import (
	"encoding/csv"
	"go-com/core/logr"
	"io"
	"os"
)

// Read 读取单个csv文件
func Read(filename string, comma rune, handler func(record []string)) {
	f, err := os.Open(filename)
	if err != nil {
		logr.L.Fatal(err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.Comma = comma
	record := make([]string, 0, 1024)
	for {
		record, err = reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			logr.L.Fatal(err)
		}
		handler(record)
	}
}

// Write 写入单个csv文件
func Write(filename string, comma rune, data [][]string) {
	// 创建csv文件
	file, err := os.Create(filename)
	if err != nil {
		logr.L.Fatal(err)
	}
	defer file.Close()

	// 将数据写入csv文件
	writer := csv.NewWriter(file)
	writer.Comma = comma
	defer writer.Flush()
	err = writer.WriteAll(data)
	if err != nil {
		logr.L.Fatal(err)
	}
}
