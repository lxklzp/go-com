package network

import (
	"go-com/core/logr"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
)

type Udp struct {
	conn *net.UDPConn
}

// Run addr格式 127.0.0.1:9601
func (u *Udp) Run(addr string) {
	// 绑定ip和端口
	var err error
	serverAddr := strings.Split(addr, ":")
	port, _ := strconv.Atoi(serverAddr[1])
	u.conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(serverAddr[0]), Port: port})
	if err != nil {
		logr.L.Fatal(err)
	}
	// panic后重新拉起服务
	defer func() {
		if err := recover(); err != nil {
			logr.L.Error(err)
		}
		u.conn.Close()
		go u.Run(addr)
	}()

	// 限制并发处理数据包的协程数目 100
	limit := make(chan bool, 100)
	var index int64
	for {
		data := make([]byte, packageMaxLen)
		// 读取UDP数据
		count, clientAddr, err := u.conn.ReadFromUDP(data)
		if err != nil {
			logr.L.Error(err)
			continue
		}
		atomic.AddInt64(&index, 1)

		limit <- true
		// 处理数据包
		go u.handle(clientAddr, data[:count], limit)
	}
}

func (u *Udp) handle(addr *net.UDPAddr, raw []byte, limit chan bool) {
	defer func() {
		<-limit
		if err := recover(); err != nil {
			logr.L.Error(err)
		}
	}()

	// 下发数据
	if _, err := u.conn.WriteToUDP(raw, addr); err != nil {
		logr.L.Error(err)
	}
}
