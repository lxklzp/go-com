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

var T tcpServer

type tcpServer struct {
	addr               string
	reset              func()
	reqDataPool        *sync.Pool                              // 请求数据的[]byte池
	request            func(req TcpUdpRequest) ([]byte, error) // 处理请求数据，并返回响应数据
	connectionCount    atomic.Int64                            // 连接数目
	maxConnectionCount int64                                   // 最大连接数目
}

func (t *tcpServer) Init(addr string, reset func(), request func(req TcpUdpRequest) ([]byte, error), reqDataMaxLen int64, maxConnectionCount int64) {
	// 初始化数据
	t.addr = addr
	t.reset = reset
	t.reqDataPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, reqDataMaxLen)
		},
	}
	t.request = request
	t.maxConnectionCount = maxConnectionCount

	// tcp服务启动
	for {
		t.run()
		logr.L.Error("[tcp] 异常关闭，等待重启...")
		time.Sleep(time.Second * 30) // tcp连接处理失败后休眠一段时间
	}
}

// 运行tcp连接处理
func (t *tcpServer) run() {
	t.reset()

	// 绑定ip和端口
	listen, err := net.Listen("tcp", t.addr)
	if err != nil {
		logr.L.Fatal(err)
	}
	defer func() {
		if err := recover(); err != nil {
			tool.ErrorStack(err)
		}
		listen.Close()
	}()

	logr.L.Debug("[tcp] 服务启动。")

	for {
		// accept阻塞，直到有新的连接
		conn, err := listen.Accept()
		if err != nil {
			logr.L.Error(err)
			continue
		}
		if !tool.AtomicIncr(&t.connectionCount, 1, t.maxConnectionCount) {
			logr.L.Info(fmt.Sprintf("[tcp] 连接数目已达到%d阈值，主动断开客户端%s连接。", t.maxConnectionCount, conn.RemoteAddr()))
			conn.Close()
			continue
		}
		logr.L.Debug(fmt.Sprintf("[tcp] 建立新的连接：%s，当前连接数：%d", conn.RemoteAddr(), t.connectionCount.Load()))
		// 处理连接
		go func() {
			defer func() {
				logr.L.Info(fmt.Sprintf("[tcp] 主动断开客户端%s连接。", conn.RemoteAddr()))
				conn.Close()
				tool.AtomicDecr(&t.connectionCount, 1, 0)
			}()
			t.connect(conn)
		}()
	}
}

// 保持并处理单个tcp连接
func (t *tcpServer) connect(conn net.Conn) {
	var err error
	for {
		if err = t.read(conn); err != nil {
			logr.L.Error(err)
			return
		}
	}
}

// 接收tcp数据
func (t *tcpServer) read(conn net.Conn) error {
	data := t.reqDataPool.Get().([]byte)
	defer func() {
		if err := recover(); err != nil {
			tool.ErrorStack(err)
		}
		t.reqDataPool.Put(data)
	}()

	count, err := conn.Read(data)
	if count == 0 {
		return fmt.Errorf("[tcp] 接收客户端%s数据长度为0，断开连接。", conn.RemoteAddr())
	}
	if err != nil {
		return err
	}
	logr.L.Debug(fmt.Sprintf("[tcp] 接收客户端%s数据，字节展示：", conn.RemoteAddr()), data[:count])
	logr.L.Debug(fmt.Sprintf("[tcp] 接收客户端%s数据，文本展示：%s", conn.RemoteAddr(), string(data[:count])))
	if err = t.handle(conn, data[:count]); err != nil {
		return err
	}
	return nil
}

// 处理单个tcp请求数据
func (t *tcpServer) handle(conn net.Conn, reqData []byte) error {
	defer func() {
		if err := recover(); err != nil {
			tool.ErrorStack(err)
		}
	}()

	// 处理请求数据，生成下发数据
	respData, err := t.request(TcpUdpRequest{
		ReqData:          reqData,
		TcpUdpConnection: TcpUdpConnection{Type: TCP, TcpConn: conn, TcpClientAddr: conn.RemoteAddr().(*net.TCPAddr)},
	})
	if err != nil {
		return err
	}

	// 下发数据
	if err = t.Send(conn, respData); err != nil {
		logr.L.Error(err)
	}
	return nil
}

// Send 下发数据
func (t *tcpServer) Send(conn net.Conn, respData []byte) error {
	// 空数据不发送
	if len(respData) == 0 {
		return nil
	}

	_, err := conn.Write(respData)
	if err != nil {
		return err
	}
	logr.L.Debug(fmt.Sprintf("[tcp] 下发客户端%s数据，字节展示：", conn.RemoteAddr()), respData)
	logr.L.Debug(fmt.Sprintf("[tcp] 下发客户端%s数据，文本展示：%s", conn.RemoteAddr(), string(respData)))
	return nil
}

// GetConnectionCount 获取当前连接数
func (t *tcpServer) GetConnectionCount() int {
	return int(t.connectionCount.Load())
}
