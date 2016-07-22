package h32

import (
	"strconv"
	"sync"
	"time"
)

// H32 Hash.
type H32 interface {
	GetUID() string
	GetOffset(key string, bucketSize, bufferSize int64) int64
	Init()
}

// h32 数据结构体.
type h32 struct {
	n int64
	sync.RWMutex
}

// Init 初始化，获取当前毫秒时间戳.
func (h *h32) Init() {
	h.Lock()
	defer h.Unlock()

	h.n = time.Now().UnixNano()
}

// GetUID 获取一个唯一KEY.
func (h *h32) GetUID() string {
	h.n++

	return strconv.FormatInt(int64(h.n), 32)
}

// GetOffset key在数组的偏移.
func (h *h32) GetOffset(key string, blockSize, bucketSize int64) int64 {
	key64, err := strconv.ParseInt(key, 32, 64)
	if err != nil {

		return 0
	}

	return key64 / bucketSize % blockSize
}

// NewHash 初始化，获取当前毫秒时间戳.
func NewHash() H32 {
	h := &h32{}
	h.Init()

	return h
}

// DefaultHash 默认获取一个KEY生成器.
var DefaultHash = NewHash()
