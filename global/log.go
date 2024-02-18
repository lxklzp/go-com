package global

import (
	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
	"go-com/config"
	"io"
	"log"
	"os"
	"strconv"
)

var Log *logrus.Logger

type logFormatter struct{}

// Format 日志格式
func (m *logFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var content = "[" + entry.Time.Format(DateTimeFormatter) + "] " + "[" + entry.Level.String() + "] " + "[" + entry.Caller.File + ":" + strconv.Itoa(entry.Caller.Line) + "] " + entry.Message + "\n"
	return []byte(content), nil
}

func InitLog(filename string) {
	Log = logrus.New()
	Log.SetLevel(logrus.DebugLevel)
	Log.SetFormatter(&logFormatter{})
	Log.SetReportCaller(true) // 记录go文件和行号信息

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
	if config.C.App.DebugMode {
		Log.SetOutput(io.MultiWriter(writer, os.Stdout)) // 输出到文件和控制台
	} else {
		Log.SetOutput(writer) // 输出到文件
	}
}
