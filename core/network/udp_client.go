package network

import (
	"github.com/pkg/errors"
	"go-com/core/logr"
	"net"
)

func UdpClientPing(addr string, reqData []byte) ([]byte, error) {
	// 根据addr解析ip和端口
	ip, port, err := ParseAddr(addr)
	if err != nil {
		logr.L.Fatal(err)
	}
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.ParseIP(ip), Port: port})
	if err != nil {
		return nil, err
	}
	// 发送数据
	count, err := socket.Write(reqData)
	logr.L.Debug("write bytes len:", count)
	if err != nil {
		return nil, errors.New("发送数据失败。")
	}

	buf := make([]byte, 1024)
	count, err = socket.Read(buf)
	logr.L.Debug("read bytes len:", count)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
