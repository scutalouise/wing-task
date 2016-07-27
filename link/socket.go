package link

import (
	"log"
)

// Hander 业务函数，当有一个请求，调用该函数.
type Hander interface {
	ServeDo(conn Connect, data [][]byte)
}

// Connect 网络请求的连接对象.
type Connect interface {
	// WriteString 写入字符串数据.
	WriteString(strs ...string) error
	// 读取指定大小数据.
	ReadSize(n int) ([]byte, error)
	// GetC 获取一个通信对象，如果网络连接关闭，获取到数据.
	GetC() chan interface{}
	// ReadOneRequest 读取一个完整的请求.
	ReadOneRequest() ([][]byte, error)
	// Close Close
	Close() error
	// Serve 开启一个服务.
	Serve()
	// StopServer 关闭服务.
	StopServer()
	// 读取数据头的数据长度.
	ReadLenLine() (int, error)
}

// Server 启动服务.
type Server interface {
	ListenAndServe() error // 监听服务.
}

// ListenAndServe 监听服务.
func ListenAndServe(addr string, call func(d interface{}), log *log.Logger) error {
	serve := &server{
		Addr:        addr,
		CallFunc:    call,
		ErrorLog:    log,
		StopMessage: make(chan interface{}, 1),
	}

	return serve.ListenAndServe()
}

// RegisterHandler 注册函数.
func RegisterHandler(cmd string, f func(conn Connect, d [][]byte)) {
	DefaultServeMux.RegisterHandler(cmd, HandlerFunc(f))
}

// DeregisterHandler 注册函数.
func DeregisterHandler(cmd string) {
	DefaultServeMux.DeregisterHandler(cmd)
}
