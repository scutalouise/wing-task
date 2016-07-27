package link

import (
	"fmt"
	"regexp"
	"sync"
)

// HandlerFunc 注册函数.
type HandlerFunc func(Connect, [][]byte)

// ServeMux 动作注册存储.
type ServeMux struct {
	sync.RWMutex                      // 锁.
	m            map[string]*muxEntry //  动作集合.
}

// muxEntry 动作信息.
type muxEntry struct {
	h   Hander // 调用函数.
	cmd string // 命令调用函数.
}

// 没有发现命令.
func notFound(conn Connect, _ [][]byte) {
	conn.WriteString("404", "not found action")
}

// exiting 退出程序中.
func exiting(conn Connect, _ [][]byte) {
	conn.WriteString("0", "server exiting")
}

// parent 命令格式正则表达式.
const parent = `^[a-zA-Z_]+[a-zA-Z0-9_]*$`

// RegisterHandler 注册动作.
func (mux *ServeMux) RegisterHandler(cmd string, handler HandlerFunc) {
	mux.Lock()
	defer mux.Unlock()

	ma, err := regexp.Match(parent, []byte(cmd))
	if err != nil {
		panic(err.Error)
	} else if ma == false {
		panic("cmd format exption")
	}

	mux.m[cmd] = &muxEntry{cmd: cmd, h: Hander(handler)}
}

// DeregisterHandler 注册动作.
func (mux *ServeMux) DeregisterHandler(cmd string) {
	mux.Lock()
	defer mux.Unlock()

	ma, err := regexp.Match(parent, []byte(cmd))
	if err != nil {
		panic(err.Error)
	} else if ma == false {
		panic("cmd format exption")
	}

	cmd = mux.getExitCmd(cmd)
	mux.m[cmd] = &muxEntry{cmd: cmd, h: exitingHandklerFunc()}
}

func (mux *ServeMux) getExitCmd(cmd string) string {

	return fmt.Sprintf("1exit_%s", cmd)
}

// notFountHandklerFunc 没有发现命令.
func notFountHandklerFunc() Hander {

	return Hander(HandlerFunc(notFound))
}

// exitingHandklerFunc 服务退出程序，注册动作调用退出提示动作.
func exitingHandklerFunc() Hander {

	return Hander(HandlerFunc(exiting))
}

// GetHandler 取出命令动作.
func (mux *ServeMux) GetHandler(cmd string) Hander {
	mux.RLock()
	defer mux.RUnlock()

	// 查询退出动作.
	_cmd := mux.getExitCmd(cmd)
	entry, ok := mux.m[_cmd]
	if ok {

		return entry.h
	}

	entry, ok = mux.m[cmd]
	if ok {

		return entry.h
	}

	return notFountHandklerFunc()
}

// ServeDo 业务action.
func (f HandlerFunc) ServeDo(conn Connect, data [][]byte) {
	f(conn, data)
}

// NewServeMux 新建一个存储命令函数.
func NewServeMux() *ServeMux {

	return &ServeMux{
		m: make(map[string]*muxEntry),
	}
}

// DefaultServeMux 默认配置.
var DefaultServeMux = NewServeMux()
