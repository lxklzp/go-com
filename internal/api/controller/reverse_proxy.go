package controller

import (
	"bytes"
	"compress/gzip"
	"github.com/gin-gonic/gin"
	"go-com/core/logr"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

/* 反向代理示例 */

var Nifi nifi

type nifi struct {
}

type processGroups struct {
	Id string `json:"id"`
}

type processGroupsSnippetRespData struct {
	Flow struct {
		ProcessGroups []processGroups `json:"processGroups"`
	} `json:"flow"`
}

type snippetDeleteRespData struct {
	Snippet struct {
		ProcessGroups map[string]interface{} `json:"processGroups"`
	} `json:"snippet"`
}

// ProxyApi nifi-api 代理
func (ctl nifi) ProxyApi(c *gin.Context) {
	urlNifi, _ := url.Parse("http://127.0.0.1")
	proxy := httputil.NewSingleHostReverseProxy(urlNifi)
	// 错误处理
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.Write([]byte(err.Error()))
	}

	proxy.Director = func(req *http.Request) {
		// group删改查鉴权
		if (strings.HasPrefix(req.URL.Path, "/nifi-api/process-groups") &&
			(req.Method == "GET" || req.Method == "PUT" || req.Method == "DELETE")) ||
			(strings.HasPrefix(req.URL.Path, "/nifi-api/flow/process-groups") && req.Method == "GET") {
		}
		req.Header = c.Request.Header
		req.Host = c.Request.Host // 使用客户端Host，避免301重定向问题
		req.URL.Scheme = "http"
		req.URL.Host = urlNifi.Host
		req.URL.RawPath = c.Request.URL.RawPath
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		ctl.readProxyRespBody(resp)
		if strings.HasPrefix(resp.Request.URL.Path, "/nifi-api/process-groups") &&
			strings.HasSuffix(resp.Request.URL.Path, "/process-groups") && resp.Request.Method == "POST" {
		} else if strings.HasPrefix(resp.Request.URL.Path, "/nifi-api/process-groups") &&
			strings.HasSuffix(resp.Request.URL.Path, "/snippet-instance") && resp.Request.Method == "POST" {
		} else if strings.HasPrefix(resp.Request.URL.Path, "/nifi-api/snippets") && resp.Request.Method == "DELETE" {
		} else if strings.HasPrefix(resp.Request.URL.Path, "/nifi-api/parameter-contexts") &&
			strings.Count(resp.Request.URL.Path, "/") == 2 && resp.Request.Method == "POST" {
		}

		return nil
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

var nifiResponseRewrite map[string]string

// 读取代理的nifi响应数据
func (ctl nifi) readProxyRespBody(resp *http.Response) []byte {
	// 生成缓存
	buffer := bytes.NewBuffer(make([]byte, 0, 4096))
	_, err := io.Copy(buffer, resp.Body)
	if err != nil {
		logr.L.Error(err)
	}
	resp.Body = io.NopCloser(buffer)

	// 解压
	var result []byte
	if resp.Header.Get("Content-Encoding") == "gzip" {
		r, err := gzip.NewReader(buffer)
		if err != nil {
			logr.L.Error(err)
			return nil
		}
		result, err = io.ReadAll(r)
		if err != nil {
			logr.L.Error(err)
			return nil
		}

		// 替换字符
		resultStr := string(result)
		for src, dst := range nifiResponseRewrite {
			resultStr = strings.Replace(resultStr, src, dst, -1)
		}
		result = []byte(resultStr)
		// 压缩
		w := gzip.NewWriter(buffer)
		if _, err = w.Write(result); err != nil {
			logr.L.Error(err)
			return nil
		}
		w.Flush()
		w.Close()
		resp.Header.Set("Content-Length", strconv.Itoa(buffer.Len()))
	} else {
		result = buffer.Bytes()

		// 替换字符
		resultStr := string(result)
		for src, dst := range nifiResponseRewrite {
			resultStr = strings.Replace(resultStr, src, dst, -1)
		}
		result = []byte(resultStr)
		buffer.Reset()
		buffer.Write(result)
		resp.Header.Set("Content-Length", strconv.Itoa(buffer.Len()))
	}

	return result
}
