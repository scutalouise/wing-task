package queue

import (
	"time"
)

// BlockSize 数据存储的数组大小.
const BlockSize = 100

// BucketSize 数据分块存储的桶.
const BucketSize = 10000

// Queue 队列接口.
type Queue interface {
	// Join 向队列中，添加一个任务.
	Join(tube, key string, value []byte) error
	// Finish 完成一个任务.
	Finish(key string, conn interface{}) bool
	// GetAndDoing 获取一个任务，修改任务状态为正在开始中.
	GetAndDoing(tube string, conn interface{}) (string, []byte, bool)
	// Exists 判定一个人是否存在, 该任务必须为未开始，正在完成中.
	Exists(key string) bool
	// RestoreOne 还原一个任务.
	RestoreOne(tube string, conn interface{}) bool
	// Usr1 添加一个通知，如果有数据，通知用户.
	Usr1(tube string, ch chan interface{}) (bool, error)
	// RestoreAll 还原一个连接对象正在做的任务进行还原.
	RestoreAll(conn interface{}) error
	// StartAndGC GC数据回收.
	StartAndGC() error
}

// READY 等待状态 RESERVED 进行中状态 DELAYED 可以删除状态.
const (
	_ uint8 = iota
	// READY 等待状态.
	READY
	// RESERVED 进行中状态.
	RESERVED
	// DELAYED 可以删除状态.
	DELAYED
)

// NewQueue 创建一个默认队列.
func NewQueue(gcTime time.Duration) Queue {
	q := &queue{
		tube: make(map[string]*li, 0),
		log:  make(map[interface{}]map[string]interface{}, 0),
		db:   make([]map[string]*job, BlockSize),
		dur:  gcTime,
	}
	q.StartAndGC()

	return q
}
