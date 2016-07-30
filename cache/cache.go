package cache

import (
	"time"
)

// BlockSize 数据存储的数组大小.
const BlockSize = 100

// BucketSize 数据分块存储的桶.
const BucketSize = 10000

// Cache 缓存接口.
type Cache interface {
	// Set 向缓存中，添加一个值.
	Set(key string, value []byte, lifespan time.Duration) error
	// Get 获取一个值.
	Get(key string) ([]byte, bool)
	// Cover 将一个已经存在的值，覆盖.
	Cover(key string, value []byte) (bool, error)
	// GetAndTimeOut 获取一个值且有时间限制.
	GetAndTimeOut(key string, time time.Duration, ch chan interface{}) ([]byte, error)
	// ClearAll 清空缓存.
	ClearAll() error
	// Delete 删除一个缓存.
	Delete(key string) bool
	// 定时回收数据.
	StartAndGC() error
}

// NewCache 新建一个缓存.
// dur GC垃圾回收周期.
func NewCache(dur time.Duration) Cache {
	c := &Block{
		dur:      dur,
		buckets:  make([]map[string]*Item, BlockSize),
		channels: make(map[string]map[chan interface{}]interface{}, 0),
	}
	c.StartAndGC()

	return c
}
