package tool

import (
	"bytes"
	"github.com/pkg/errors"
	"go-com/core/logr"
	"io"
	"net/http"
	"strconv"
	"time"
)

func httpReqResp(req *http.Request, url string, param interface{}) ([]byte, error) {
	// 请求
	client := &http.Client{}
	client.Timeout = time.Minute
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	// 处理返回结果
	result, err := io.ReadAll(resp.Body)
	logr.L.Debugf("请求 %s:%s\n响应 [%d] %s", url, param, resp.StatusCode, result)
	if err != nil {
		return nil, err
	} else if (resp.StatusCode < http.StatusOK) || (resp.StatusCode > 299) {
		return result, errors.New("服务器异常，响应码：" + strconv.Itoa(resp.StatusCode))
	} else {
		return result, nil
	}
}

func Get(url string, param map[string]string, header map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// header头
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	// 参数
	if len(param) > 0 {
		query := req.URL.Query()
		for k, v := range param {
			query.Add(k, v)
		}
		req.URL.RawQuery = query.Encode()
	}

	return httpReqResp(req, url, param)
}

// Post 请求参数格式：json
func Post(url string, param []byte, header map[string]string, method string) ([]byte, error) {
	body := bytes.NewBuffer(param)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	// header头
	req.Header.Set("Content-Type", "application/json")
	if len(header) > 0 {
		for k, v := range header {
			req.Header.Set(k, v)
		}
	}

	return httpReqResp(req, url, param)
}
