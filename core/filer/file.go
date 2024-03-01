package filer

import (
	"archive/zip"
	"bufio"
	"bytes"
	"go-com/core/logr"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ScanSuffixFile 递归遍历文件夹，处理指定后缀的文件名
func ScanSuffixFile(dir string, suffix string, handler func(filename string) error) error {
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 跳过目录自身
		if path == dir {
			return nil
		}
		filename := info.Name()
		if info.IsDir() {
			return nil
		} else if strings.HasSuffix(filename, suffix) {
			return handler(filename)
		}
		return err
	})
}

// Exist 文件/文件夹是否存在
func Exist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// Zip 压缩文件夹
func Zip(srcDir string, dstDir string, zipFilename string) error {
	if !Exist(dstDir) {
		err := os.MkdirAll(dstDir, os.ModePerm)
		return err
	}
	// 创建新的压缩文件
	archive, err := os.Create(dstDir + "/" + zipFilename)
	if err != nil {
		return err
	}
	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()
	err = filepath.Walk(srcDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		logr.L.Debug("walk", path)
		// 跳过目录自身
		if path == srcDir {
			return nil
		}
		// 获取zip包中的相对路径 比如要压缩的目录是/tmp/tozip
		// 要压缩的文件是/tmp/tozip/tozip.file
		// 则得到的zipPath = tozip.file
		// 保证压缩后文件目录结构和之前是一样的
		// 如果需要使用新的目录，可以根据需要自定义
		zipPath := path[len(srcDir)+1:]
		if info.IsDir() {
			zipPath += "/"
		}
		w, err := zipWriter.Create(zipPath)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fr, err := os.Open(path)
		defer fr.Close()
		if err != nil {
			return err
		}
		_, err = io.Copy(w, fr)
		if err != nil {
			return err
		}
		return nil
	})
	// 在这里读取新的zip文件可能会出问题
	// 除非把上面的defer zipWriter.Close()去掉，然后在这里先执行zipWriter.Close()在读取
	return err
}

// Unzip 解压
func Unzip(zipData []byte, destDir string) error {
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return err
	}
	for _, f := range zipReader.File {
		err = writeUnzipFile(f, destDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeUnzipFile(f *zip.File, destDir string) error {
	fName := f.Name
	destPath := filepath.Join(destDir, fName)
	// 处理zip包含多层文件目录的情况
	if f.FileInfo().IsDir() {
		return os.MkdirAll(destPath, os.ModePerm)
	}
	// 创建要写入的文件
	fw, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer fw.Close()
	fr, err := f.Open()
	if err != nil {
		return err
	}
	defer fr.Close()
	_, err = io.Copy(fw, fr)
	return err
}

// ReadLine 按行读取数据
func ReadLine(filename string, handler func(line string) error) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(f)
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		err = handler(line)
		if err != nil {
			return err
		}
	}
	return nil
}

func CopyFile(dstName, srcName string) error {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstName, os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}
