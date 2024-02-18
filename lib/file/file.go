package file

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/xuri/excelize/v2"
	"go-com/global"
	"io"
	"os"
	"strings"
)

// ExcelRead 读取示例
func ExcelRead(filename string, sheet string, from int) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		global.Log.Error(err)
		return
	}
	defer f.Close()

	rows, err := f.GetRows(sheet)
	if err != nil {
		global.Log.Error(err)
		return
	}
	rows = rows[from:]

	for _, row := range rows {
		fmt.Println(row)
	}
}

// CsvTemplateRead 遍历模板目录，获取csv模板标题信息
func CsvTemplateRead(path string, comma rune) {
	dirs, err := os.ReadDir(path)
	if err != nil {
		global.Log.Panic(err)
	}
	var filename string
	var content []byte
	for _, v := range dirs {
		filename = v.Name()
		if strings.HasSuffix(filename, ".csv") {
			content, err = os.ReadFile(path + "/" + filename)
			if err != nil {
				global.Log.Panic(err)
			}
			title := strings.Split(strings.TrimSpace(string(content)), string(comma))
			fmt.Println(title)
		}
	}
}

// CsvWrite csv写入 filename格式 xxx.csv
func CsvWrite(filename string, comma rune, data [][]string) {
	// 创建csv文件
	file, err := os.Create(filename)
	if err != nil {
		global.Log.Fatal(err)
	}
	defer file.Close()

	// 将数据写入csv文件
	writer := csv.NewWriter(file)
	writer.Comma = comma
	defer writer.Flush()
	err = writer.WriteAll(data)
	if err != nil {
		global.Log.Fatal(err)
	}
}

// Zip 压缩文件夹 dis格式 xx.zip
func Zip(src string, dst string) {
	var err error
	zipFile, err := os.Create(dst)
	if err != nil {
		global.Log.Panic(err)
	}
	defer zipFile.Close()

	// 压缩文件写句柄
	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	// 压缩csv缓存目录
	dirs, err := os.ReadDir(src)
	if err != nil {
		global.Log.Panic(err)
	}
	var csvName string
	for _, v := range dirs {
		csvName = v.Name()
		if strings.HasSuffix(v.Name(), ".csv") {
			// 压缩文件
			writer, _ := archive.Create(csvName)
			file, _ := os.Open(src + "/" + v.Name())
			io.Copy(writer, file)
			file.Close()
		}
	}
}

// ReadLine 按行读取数据
func ReadLine(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		global.Log.Fatal(err)
	}
	reader := bufio.NewReader(f)
	for {
		var line string
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			global.Log.Error(err)
		} else {
			fmt.Println(line)
		}
	}
}
