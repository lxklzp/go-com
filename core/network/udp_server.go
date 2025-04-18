package network

import (
	"fmt"
	"go-com/core/logr"
	"go-com/core/tool"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

var U udpServer

type udpServer struct {
	addr           string
	conn           *net.UDPConn
	reset          func()
	reqDataPool    *sync.Pool                              // 请求数据的[]byte池
	request        func(req TcpUdpRequest) ([]byte, error) // 处理请求数据，并返回响应数据
	handleCount    atomic.Int64                            // 连接数目
	maxHandleCount int64                                   // 最大连接数目
}

func (u *udpServer) Init(addr string, reset func(), request func(req TcpUdpRequest) ([]byte, error), reqDataMaxLen int64, maxHandleCount int64) {
	// 初始化数据
	u.addr = addr
	u.reset = reset
	u.reqDataPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, reqDataMaxLen)
		},
	}
	u.request = request
	u.maxHandleCount = maxHandleCount

	// udp服务启动
	for {
		u.run()
		logr.L.Error("[udp] 异常关闭，等待重启...")
		time.Sleep(time.Second * 30) // udp连接处理失败后休眠一段时间
	}
}

// 运行udp连接处理
func (u *udpServer) run() {
	u.reset()

	// 根据addr解析ip和端口
	ip, port, err := ParseAddr(u.addr)
	if err != nil {
		logr.L.Fatal(err)
	}
	// 绑定ip和端口
	u.conn, err = net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(ip), Port: port})
	if err != nil {
		logr.L.Fatal(err)
	}
	defer func() {
		if err := recover(); err != nil {
			tool.ErrorStack(err)
		}
		u.conn.Close()
	}()

	logr.L.Debug("[udp] 服务启动。")

	// 处理udp请求数据
	for {
		data := u.reqDataPool.Get().([]byte)
		// 读取UDP数据
		count, clientAddr, err := u.conn.ReadFromUDP(data)
		if err != nil {
			logr.L.Error(err)
			continue
		}

		logr.L.Debug(fmt.Sprintf("[udp] 接收客户端%s数据，字节展示：", clientAddr), data[:count])
		logr.L.Debug(fmt.Sprintf("[udp] 接收客户端%s数据，文本展示：%s", clientAddr, string(data[:count])))

		if !tool.AtomicIncr(&u.handleCount, 1, u.maxHandleCount) {
			logr.L.Info(fmt.Sprintf("[udp] 同时处理的客户端数目已达到%d阈值，丢弃客户端%s数据。", u.maxHandleCount, clientAddr))
			continue
		}

		// 处理单个udp请求数据
		go func() {
			defer func() {
				u.reqDataPool.Put(data)
				tool.AtomicDecr(&u.handleCount, 1, 0)
			}()
			u.handle(clientAddr, data[:count])
		}()
	}
}

// 处理单个udp请求数据
func (u *udpServer) handle(clientAddr *net.UDPAddr, reqData []byte) {
	defer func() {
		if err := recover(); err != nil {
			tool.ErrorStack(err)
		}
	}()

	// 处理请求数据，生成下发数据
	respData, err := u.request(TcpUdpRequest{
		ReqData:          reqData,
		TcpUdpConnection: TcpUdpConnection{Type: UDP, UdpClientAddr: clientAddr},
	})
	if err != nil {
		logr.L.Error(err)
		return
	}

	// 下发数据
	if err = u.Send(clientAddr, respData); err != nil {
		logr.L.Error(err)
	}
}

// Send 下发数据
func (u *udpServer) Send(clientAddr *net.UDPAddr, respData []byte) error {
	// 空数据不发送
	if len(respData) == 0 {
		return nil
	}

	if _, err := u.conn.WriteToUDP(respData, clientAddr); err != nil {
		return err
	}
	logr.L.Debug(fmt.Sprintf("[udp] 下发客户端%s数据，字节展示：", clientAddr), respData)
	logr.L.Debug(fmt.Sprintf("[udp] 下发客户端%s数据，文本展示：%s", clientAddr, string(respData)))
	return nil
}
