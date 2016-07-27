package link

import (
	"bufio"
	"errors"
	"log"
	"net"
	"strconv"
	"sync"
	"time"
	"bytes"
	"fmt"
	"io"
)

// server 服务结构体.
type server struct {
	Addr        string            // 监听端口与地址.
	CallFunc    func(interface{}) // 当客户端网络断开的时候，调用该函数.
	ErrorLog    *log.Logger       // 日志记录对象.
	StopMessage chan interface{}
}

// ListenAndServe 监听端口与启动服务，服务不断的建立连接与读取数据.
func (srv *server) ListenAndServe() error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		panic(err)
	}

	ac := make(chan interface{}, 1)
	defer func() {
		ln.Close()
	}()

	go srv.Serve(ln, ac)
	select {
	case <-srv.StopMessage:

		return nil
	case err := <-ac:
		if e, ok := err.(error); ok {

			return e
		}

		return nil
	}
}

func (srv *server) StopServer() {
	srv.StopMessage <- nil
}

// Serve 建立服务，监听建立网络连接同时读取数据.
func (srv *server) Serve(ln net.Listener, ac chan interface{}) {
	var tempDelay time.Duration
	defer func() {
		if e := recover(); e != nil {
			srv.logf("tcp: read error: %v; retrying in %v", e, tempDelay)
			ac <- e
		}
	}()

	for {
		rw, e := ln.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				srv.logf("tcp: read error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			ac <- e
		}
		tempDelay = 0
		c := srv.NewConn(rw)
		go c.Serve()
	}
}

// logf 记录错误日志.
func (srv *server) logf(format string, args ...interface{}) {
	if srv.ErrorLog != nil {
		srv.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

// NewConn 新建连接.
func (srv *server) NewConn(rw net.Conn) Connect {

	conn := &connect{
		conn: rw,
		buf:  bufio.NewReader(rw),
		srv:  srv,
		C:    make(chan interface{}, 2),
	}

	return conn
}

// NewConnect 一个连接请求.
func NewConnect(rw net.Conn) Connect {
	conn := &connect{
		conn: rw,
		buf:  bufio.NewReader(rw),
		srv:  nil,
		C:    make(chan interface{}, 2),
	}

	return conn
}

// connect 连接结构体.
type connect struct {
	sync.RWMutex          // 读写锁.
	conn net.Conn         // 网络连接，接口对象.
	buf  *bufio.Reader    // buf读取缓存.
	srv  *server          // 服务结构对象.
	C    chan interface{} // 通知网络连接是否断开.
}

// Serve 网络连接服务，不断读取数据.
func (linker *connect) Serve() {
	for {
		b, err := linker.ReadOneRequest()
		if err != nil {
			linker.logf("tcp: connect error: %v; retrying in %v", err, linker.conn.RemoteAddr())
			if err == io.EOF {
				// 如果连接关闭，或者客户端关闭了连接.
				linker.Close()
				linker.srv.CallFunc(linker)
				linker.C <- nil
				break
			}
		}
		if b != nil {
			handler := DefaultServeMux.GetHandler(string(b[0]))
			go handler.ServeDo(linker, b)
		}
	}
}

// logf 日志记录.
func (linker *connect) logf(format string, args ...interface{}) {
	linker.srv.logf(format, args...)
}

// GetC 获取通信数据，如果连接关闭，获取到数据.
func (linker *connect) GetC() chan interface{} {

	return linker.C
}

// WriteString 写入字符串数据.
func (linker *connect) WriteString(strs ...string) (err error) {
	b := format(strs...)
	defer func() {
		if e := recover(); e != nil {
			if err, ok := e.(error); ok {
				linker.logf("tcp: write error: %v; retrying in %v", err, linker.conn.RemoteAddr())
			}
		}
	}()
	l := len(b)
	var count, n int
	for count < l {
		n, err = linker.conn.Write(b[count:])
		if err != nil {
			return nil
		}
		count += n
	}

	return nil
}

// StopServer 停止服务.
func (linker *connect) StopServer() {
	linker.srv.StopServer()
}

// 获取一个完整的请求.
func (linker *connect) ReadOneRequest() (b [][]byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			if er, ok := e.(error); ok {
				err = er
			}
		}
	}()

	_, err = linker.buf.ReadSlice('*')
	if err != nil {

		return nil, err
	}

	len, err := linker.ReadLenLine()
	if err != nil {

		return nil, err
	}

	if len < 1 {

		return nil, nil
	}

	b = make([][]byte, len)
	var dataSize int
	for k := range b {
		_, err = linker.buf.ReadSlice('$')
		if err != nil {

			return nil, err
		}

		dataSize, err = linker.ReadLenLine()
		if err != nil {

			return nil, err
		}

		b[k], err = linker.ReadSize(dataSize)
		if err != nil {

			return nil, err
		}
	}

	return b, nil
}

// ReadLenLine 读取一个完整行数据.
func (linker *connect) ReadLenLine() (int, error) {
	b, err := linker.buf.ReadSlice('\n')
	if err != nil {
		return 0, err
	}
	l := len(b)
	if l <= 1 {
		return 0, errors.New("-1")
	}
	b = b[:l - 1]

	l, err = strconv.Atoi(string(b))

	return l, err

}

// read 读取数据.
func (linker *connect) read(b []byte) (int, error) {

	return linker.buf.Read(b)
}

// ReadSize 读取指定大小数据.
func (linker *connect) ReadSize(size int) ([]byte, error) {
	if size < 1 {

		return nil, nil
	}

	b := make([]byte, size)
	var s int
	for s < size {
		n, err := linker.read(b[s:])

		if err != nil {

			return nil, err
		}
		s += n
	}

	return b, nil
}

// Close.
func (linker *connect) Close() error {
	return linker.conn.Close()
}

// format 格式数据.
func format(strs ...string) ([]byte) {
	buf := bytes.NewBuffer(nil)
	fmt.Fprintf(buf, "*%d\n", len(strs))
	for _, str := range strs {
		fmt.Fprintf(buf, "$%d\n%s\n", len(str), str)
	}

	return buf.Bytes()
}
