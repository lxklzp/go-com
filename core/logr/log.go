package logr

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"go-com/config"
	"io"
	"log"
	"os"
)

var L *logrus.Logger

type logFormatter struct{}

// Format 日志格式
func (m *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var content = fmt.Sprintf("[%s] [%s] [%s:%d] %s\n", entry.Time.Format(config.DateTimeFormatter), entry.Level, entry.Caller.File, entry.Caller.Line, entry.Message)
	return []byte(content), nil
}

type logFormatterEmpty struct{}

// Format 日志格式
func (m *logFormatterEmpty) Format(entry *logrus.Entry) ([]byte, error) {
	return []byte(entry.Message + "\n"), nil
}

func InitLog(filename string) {
	L = NewLog(filename, true)
}

func NewLog(filename string, format bool) *logrus.Logger {
	L := logrus.New()
	if config.C.App.DebugMode {
		L.SetLevel(logrus.DebugLevel)
	} else {
		L.SetLevel(logrus.InfoLevel)
	}
	if format {
		L.SetFormatter(&logFormatter{})
		L.SetReportCaller(true) // 记录go文件和行号信息
	} else {
		L.SetFormatter(&logFormatterEmpty{})
	}

	// 创建日志目录
	path := config.RuntimePath + "/log"
	if err := os.MkdirAll(path, 0777); err != nil {
		log.Fatal(err)
	}

	// 日志文件写入和分割
	writer := &lumberjack.Logger{
		Filename:  path + "/" + filename + ".log",
		MaxSize:   100,
		MaxAge:    2,
		LocalTime: true,
	}
	L.SetOutput(io.MultiWriter(writer, os.Stdout)) // 输出到文件和控制台
	return L
}
