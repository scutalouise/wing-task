package h32

import (
	"strconv"
	"sync"
	"time"
	"crypto/md5"
	"encoding/hex"
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
	h.Lock()
	defer h.Unlock()

	h.n++
	return strconv.FormatInt(h.n, 32)
}

// GetOffset key在数组的偏移.
func (h *h32) GetOffset(key string, blockSize, bucketSize int64) int64 {
	h32 := convertMD5(key)
	hash, _ := strconv.ParseInt(h32[:8], 32, 64)

	i := hash / bucketSize % blockSize

	return i
}

// NewHash 初始化，获取当前毫秒时间戳.
func NewHash() H32 {
	h := &h32{}
	h.Init()

	return h
}

// convertMD5 加密
func convertMD5(str string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(str))

	return hex.EncodeToString(md5Ctx.Sum(nil))
}

// DefaultHash 默认获取一个KEY生成器.
var DefaultHash = NewHash()
