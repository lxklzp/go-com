package gw

import (
	"github.com/gin-gonic/gin"
	"go-com/config"
	"go-com/core/tool"
	"mime/multipart"
	"os"
	"time"
)

// Upload 文件上传
func Upload(c *gin.Context) (map[string]interface{}, error) {
	var err error
	// 创建上传文件夹
	realPath := config.C.App.PublicPath
	relativePath := "/" + time.Now().Format(config.MonthNumberFormatter)
	realPath += relativePath
	if err = os.MkdirAll(realPath, 0755); err != nil {
		return nil, err
	}

	// 处理上传文件
	var f *multipart.FileHeader
	if f, err = c.FormFile("file"); err != nil {
		return nil, err
	}

	filename := "/" + time.Now().Format(config.DateTimeNumberFormatter) + "_" + f.Filename
	realPath += filename
	relativePath += filename
	if err = c.SaveUploadedFile(f, realPath); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"relative_path": relativePath,
		"size":          tool.FormatFileSize(f.Size),
	}, nil
}
