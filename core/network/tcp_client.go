package network

import (
	"github.com/pkg/errors"
	"go-com/core/logr"
	"net"
)

func TcpClientPing(addr string, reqData []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	count, err := conn.Write(reqData)
	logr.L.Debug("write bytes len:", count)
	if err != nil {
		return nil, errors.New("发送数据失败。")
	}

	buf := make([]byte, 1024)
	count, err = conn.Read(buf)
	logr.L.Debug("read bytes len:", count)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
