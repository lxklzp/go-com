package tool

import (
	"fmt"
	"github.com/pkg/errors"
	"go-com/config"
	"go-com/core/logr"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"unicode"
)

// ExitNotify 监听退出信号，关闭系统资源
func ExitNotify(close func()) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		for s := range ch {
			switch s {
			case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL:
				close()
				logr.L.Info("关闭系统")
				os.Exit(0)
			}
		}
	}()
}

// ErrorStack error返回错误栈信息
func ErrorStack(err interface{}) string {
	var msg string
	switch err.(type) {
	case error:
		msg = fmt.Sprintf("%+v", errors.WithStack(err.(error)))
	default:
		msg = fmt.Sprintf("%+v", err)
	}

	logr.L.Error(msg)
	return msg
}

type ResponseData struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func RespData(code int, message string, data interface{}) ResponseData {
	return ResponseData{Code: code, Message: message, Data: data}
}

func CamelToSepName(field string, sep rune) string {
	if field == "" {
		return ""
	}
	var buffer []rune
	for i, r := range []rune(field) {
		if unicode.IsUpper(r) {
			if i != 0 {
				buffer = append(buffer, sep)
			}
			buffer = append(buffer, unicode.ToLower(r))
		} else {
			buffer = append(buffer, r)
		}
	}
	return string(buffer)
}

func SepNameToCamel(field string, isUcFirst bool) string {
	if field == "" {
		return ""
	}
	name := strings.ReplaceAll(cases.Title(language.English).String(strings.ReplaceAll(strings.ToLower(field), "_", " ")), " ", "")
	if isUcFirst {
		return name
	}
	return strings.ToLower(name[:1]) + name[1:]
}

const LetterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = LetterBytes[rand.Intn(62)]
	}
	return string(b)
}

// FormatFileSize 格式化文件大小
func FormatFileSize(size int64) string {
	if size >= config.GB {
		return fmt.Sprintf("%.2f Gb", float64(size)/float64(config.GB))
	} else if size >= config.MB {
		return fmt.Sprintf("%.2f Mb", float64(size)/float64(config.MB))
	} else if size >= config.KB {
		return fmt.Sprintf("%.2f Kb", float64(size)/float64(config.KB))
	} else {
		return fmt.Sprintf("%d B", size)
	}
}
