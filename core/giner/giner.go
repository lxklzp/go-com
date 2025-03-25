package giner

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/tool"
	"io"
	"mime/multipart"
	"os"
	"time"
)

type UploadFileSaveInfo struct {
	RelativePath string `json:"relative_path"`
	Size         string `json:"size"`
}

// UploadFileSave 文件上传保存
func UploadFileSave(c *gin.Context) (UploadFileSaveInfo, error) {
	var err error
	var info UploadFileSaveInfo
	// 创建上传文件夹
	realPath := config.C.App.PublicPath
	relativePath := "/" + time.Now().Format(config.MonthNumberFormatter)
	realPath += relativePath
	if err = os.MkdirAll(realPath, 0755); err != nil {
		return info, err
	}

	// 处理上传文件
	var f *multipart.FileHeader
	if f, err = c.FormFile("file"); err != nil {
		return info, err
	}

	filename := "/" + time.Now().Format(config.DateTimeNumberFormatter) + "_" + f.Filename
	realPath += filename
	relativePath += filename
	if err = c.SaveUploadedFile(f, realPath); err != nil {
		return info, err
	}

	info.RelativePath = relativePath
	info.Size = tool.FormatFileSize(f.Size)
	return info, nil
}

// UploadFileRead 文件上传读取
func UploadFileRead(c *gin.Context) ([]byte, error) {
	// 读取文件内容
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return nil, err
	}
	file, err := fileHeader.Open()
	defer file.Close()
	if err != nil {
		return nil, err
	}
	dataJson, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return dataJson, err
}
