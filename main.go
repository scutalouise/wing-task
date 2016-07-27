package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"./cache"
	"./h32"
	"./link"
	"./queue"
)

// Config 配置系统.
type Config struct {
	Address  string
	Log      *log.Logger
	Filename string
}

// DefaultQueue 队列对象.
var DefaultQueue = queue.NewQueue(time.Minute * 10)

// DefaultH32 默认.
var DefaultH32 = h32.DefaultHash

// DefaultCache 缓存对象.
var DefaultCache = cache.NewCache(time.Minute * 10)

// DefaultConfig 默认配置.
var DefaultConfig = NewConfig(":8989", "/Users/liaozhouping/Desktop/task.log")

// main 主函数.
func main() {
	var cmd string
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	} else {
		cmd = ""
	}
	cmd = strings.ToLower(cmd)
	switch cmd {
	case "start":
		Start()
	case "server": // 运行后台守护进程.
		StartServer()
	case "stop":
		Stop()
	default:
		fmt.Println("支持start|stop命令")
	}
}

// Start 开启服务.
func Start() {

	// 检查服务.
	conn, err := net.DialTimeout("tcp", DefaultConfig.Address, time.Second*3)
	if err == nil {
		linker := link.NewConnect(conn)
		defer linker.Close()
		linker.WriteString("Status")
		data, err := linker.ReadOneRequest()
		if err == nil {
			if len(data) > 0 && string(data[0]) == "1" {
				fmt.Println("服务已经启动")
				return
			}
		}
	}

	// 创建后台守护进程.
	if os.Getpid() != 1 {
		filePath, _ := filepath.Abs(os.Args[0])
		os.Args[1] = "server"
		cmd := exec.Command(filePath, os.Args[1:]...)
		cmd.Stdin = nil
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
		return
	}
}

// Stop 关闭服务.
func Stop() {
	conn, err := net.DialTimeout("tcp", DefaultConfig.Address, time.Second*3)
	if err == nil {
		linker := link.NewConnect(conn)
		defer linker.Close()
		linker.WriteString("StopServer")
		data, err := linker.ReadOneRequest()
		if err == nil {
			if len(data) > 0 && string(data[0]) == "1" {
				return
			}
		}
	}

	fmt.Println("没有启动服务")
}

// StartServer 启动服务.
func StartServer() {
	DefaultConfig.Init()
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("启动服务")
	//// 注册动作.
	// AddJob 向队列添加任务.
	link.RegisterHandler("AddJob", AddJob)
	// GetReturn 获取任务完成后的结果.
	link.RegisterHandler("GetReturn", GetReturn)
	// Usr1 队列中，添加通知.
	link.RegisterHandler("Usr1", Usr1)
	// GetJob 获取任务.
	link.RegisterHandler("GetJob", GetJob)
	// SetReturn 设置任务完成结果.
	link.RegisterHandler("SetReturn", SetReturn)
	// StopServer 关闭服务.
	link.RegisterHandler("StopServer", StopServer)
	// Status 获取服务状态.
	link.RegisterHandler("Status", Status)
	// 启动网络服务.
	err := link.ListenAndServe(DefaultConfig.Address, EOF, DefaultConfig.Log)
	// 服务退出，一些注册动作不能继续使用
	logf(err)
	fmt.Println("完成退出")
}

// AddJob 添加任务.
func AddJob(conn link.Connect, d [][]byte) {
	err := addJob(conn, d)
	logf(err)
}

// NewConfig 创建默认配置.
func NewConfig(addr, logFilename string) *Config {
	c := &Config{
		Address:  addr,
		Filename: logFilename,
	}

	return c
}

// Init 初始化文件日志.
func (conf *Config) Init() {
	var l *log.Logger
	if conf.Filename != "" {
		conf.Filename, _ = filepath.Abs(conf.Filename)
		dir := filepath.Dir(conf.Filename)
		if !isDirExists(dir) {
			err := os.MkdirAll(dir, 0666)
			if err != nil {
				fmt.Println("Fail to find", err.Error(), "cServer start Failed")
				os.Exit(1)
			}
		}
		logFile, logErr := os.OpenFile(conf.Filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
		if logErr != nil {
			fmt.Println("Fail to find", logErr.Error(), "cServer start Failed")
			os.Exit(1)
		}
		l = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)
	}
	conf.Log = l
}

// isDirExists 判定目录是否存在.
func isDirExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	}
	return fi.IsDir()
}

// GetReturn 获取数据.
func GetReturn(conn link.Connect, d [][]byte) {
	err := getReturn(conn, d)
	logf(err)
}

// StopServer 停止服务.
func StopServer(conn link.Connect, d [][]byte) {
	conn.StopServer()
	conn.WriteString("1", "暂停服务中")
}

// Status 服务状态.
func Status(conn link.Connect, d [][]byte) {
	conn.WriteString("1", "运行中")
}

// Usr1 有数据通知.
func Usr1(conn link.Connect, d [][]byte) {
	err := usr1(conn, d)
	logf(err)
}

// GetJob 获取任务对象.
func GetJob(conn link.Connect, d [][]byte) {
	err := getJob(conn, d)
	logf(err)
}

// SetReturn 设置数据.
func SetReturn(conn link.Connect, d [][]byte) {
	err := setReturn(conn, d)
	logf(err)
}

func getReturn(conn link.Connect, d [][]byte) (err error) {
	l := len(d)
	if l < 2 {
		return conn.WriteString("405", "请传key")
	}

	timeout := time.Minute
	if d[2] != nil {
		tmp, err := strconv.Atoi(string(d[2]))
		if err != nil {
			timeout = time.Second * time.Duration(tmp)
		}
	}
	key := string(d[1])
	var val []byte
	ok := DefaultQueue.Exists(key)
	if ok {
		val, err = DefaultCache.GetAndTimeOut(key, timeout, conn.GetC())
		if err != nil {
			if err.Error() == "timeout" {
				return conn.WriteString("408", "超时")
			} else if err.Error() == "EOF" {
				return err
			}
			SystemERR(conn, err)
			return err
		}
		return conn.WriteString("1", "成功", string(val))
	}
	val, ok, err = DefaultCache.Get(key)
	if err != nil {
		SystemERR(conn, err)
		return err
	} else if ok {
		return conn.WriteString("1", "成功", string(val))
	}
}

func addJob(conn link.Connect, d [][]byte) (err error) {
	l := len(d)
	if l < 3 {
		return conn.WriteString("405", "参数错误")
	}
	d = d[:3]
	key := DefaultH32.GetUID()
	err = DefaultQueue.Join(string(d[1]), key, d[2])
	if err != nil {
		SystemERR(conn, err)
		return
	}
	return conn.WriteString("1", "成功", key)
}

func usr1(conn link.Connect, d [][]byte) (err error) {
	l := len(d)
	if l < 2 {
		return conn.WriteString("405", "参数错误")
	}
	d = d[:2]
	ok, err := DefaultQueue.Usr1(string(d[1]), conn.GetC())
	if err != nil {
		if err.Error() == "EOF" {
			return
		}
		SystemERR(conn, err)
		return
	} else if ok {
		return conn.WriteString("1", "成功")
	}
	SystemERR(conn, err)
	return
}

func getJob(conn link.Connect, d [][]byte) (err error) {
	l := len(d)
	if l < 2 {
		return conn.WriteString("405", "参数错误")
	}
	d = d[:2]
	key, val, err := DefaultQueue.GetAndDoing(string(d[1]), conn)
	if err != nil {
		SystemERR(conn, err)
		return
	} else if key == "" || val == nil {
		return conn.WriteString("0", "没有拿到")
	}

	return conn.WriteString("1", "成功", key, string(val))
}

func setReturn(conn link.Connect, d [][]byte) (err error) {
	l := len(d)
	if l < 3 {
		return conn.WriteString("405", "参数错误")
	}
	d = d[:3]
	key := string(d[1])
	ok, err := DefaultQueue.Finish(key, conn)
	if err != nil {
		SystemERR(conn, err)
		return
	} else if !ok {
		return conn.WriteString("404", "不存在")
	}
	err = DefaultCache.Set(key, d[2], time.Minute*10)
	if err != nil {
		SystemERR(conn, err)
		return
	}
	return conn.WriteString("1", "成功")
}

// EOF 连接断开回调函数.
func EOF(conn interface{}) {
	err := DefaultQueue.RestoreAll(conn)
	logf(err)
}

// SystemERR 系统异常.
func SystemERR(conn link.Connect, err error) {
	err = conn.WriteString("-1", "系统异常")
	logf(err)
}

func logf(err error) {
	if err != nil && DefaultConfig.Log != nil {
		DefaultConfig.Log.Printf("tcp: write error: %v", err)
	}
}
