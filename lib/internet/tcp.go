package internet

import (
	"go-com/global"
	"io"
	"net"
)

var packageMaxLen = global.KB

type Tcp struct {
}

// Run addr格式 127.0.0.1:9601
func (t *Tcp) Run(addr string) {
	// 绑定ip和端口
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		global.Log.Fatal(err)
	}
	// panic后重新拉起服务
	defer func() {
		if err := recover(); err != nil {
			global.Log.Error(err)
		}
		listen.Close()
		go t.Run(addr)
	}()

	// 限制同时连接的tcp数目 100
	limit := make(chan bool, 100)
	for {
		// accept阻塞，直到有新的连接
		conn, err := listen.Accept()
		if err != nil {
			global.Log.Error(err)
			continue
		}
		limit <- true
		// 处理连接
		go t.connect(conn, limit)
	}
}

func (t *Tcp) connect(conn net.Conn, limit chan bool) {
	defer func() {
		if err := recover(); err != nil {
			global.Log.Error(err)
		}
		conn.Close()
		<-limit
	}()

	for {
		data := make([]byte, packageMaxLen)
		count, err := conn.Read(data)
		if err != nil && err != io.EOF {
			global.Log.Error(err)
			return
		}
		if count > 10 {
			t.handle(conn, data[:count])
		}
	}
}

func (t *Tcp) handle(conn net.Conn, raw []byte) {
	defer func() {
		if err := recover(); err != nil {
			global.Log.Error(err)
		}
	}()

	// 下发数据
	if _, err := conn.Write(raw); err != nil {
		global.Log.Error(err)
	}
}
