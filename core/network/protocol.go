package network

import (
	"net"
)

const (
	TCP = "tcp"
	UDP = "udp"
)

// TcpUdpRequest 请求
type TcpUdpRequest struct {
	ReqData []byte
	TcpUdpConnection
}

// TcpUdpConnection 连接
type TcpUdpConnection struct {
	Type          string
	TcpConn       net.Conn
	TcpClientAddr *net.TCPAddr
	UdpClientAddr *net.UDPAddr
}

// GetIpPort 解析ip和端口
func (c *TcpUdpConnection) GetIpPort() (string, int) {
	var ip string
	var port int
	switch c.Type {
	case TCP:
		ip = c.TcpClientAddr.IP.String()
		port = c.TcpClientAddr.Port
	case UDP:
		ip = c.UdpClientAddr.IP.String()
		port = c.UdpClientAddr.Port
	}
	return ip, port
}
