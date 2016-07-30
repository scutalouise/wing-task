package cache

import (
	"errors"
	"sync"
	"time"

	"../h32"
)

// Block 大盒子.
type Block struct {
	sync.RWMutex                                         // 数据锁.
	dur      time.Duration                               // GC垃圾回收周期.
	buckets  []map[string]*Item                          // 原始数据.
	channels map[string]map[chan interface{}]interface{} // 注册事件, 用户订阅指定的keycache，当key数据存在时候，则通知订阅者.
}

// Item 数据存储项.
type Item struct {
	value      []byte        // 数据.
	createTime time.Time     // 创建时间.
	lifespan   time.Duration // 生命周期.
}

// Set 添加一个值.
func (box *Block) Set(key string, value []byte, lifespan time.Duration) error {
	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	box.Lock()
	defer box.Unlock()

	bucket := box.buckets[off]
	if bucket == nil {
		bucket = make(map[string]*Item, 10)
		box.buckets[off] = bucket
	}
	if channels, ok := box.channels[key]; ok {
		for channel := range channels {
			channel <- nil
		}
		delete(box.channels, key)
	}

	bucket[key] = &Item{
		value:      value,
		createTime: time.Now(),
		lifespan:   lifespan,
	}

	return nil
}

// Cover 覆盖一个值.
func (box *Block) Cover(key string, value []byte) (bool, error) {
	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	box.Lock()
	defer box.Unlock()

	if bucket := box.buckets[off]; bucket != nil {
		if itm, ok := bucket[key]; ok {
			itm.value = value
			itm.createTime = time.Now()

			return true, nil
		}
	}
	delete(box.channels, key)

	return false, nil
}

// Get 获取一个值.
func (box *Block) Get(key string) ([]byte, bool) {
	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	box.Lock()
	defer box.Unlock()

	if bucket := box.buckets[off]; bucket != nil {
		if itm, ok := bucket[key]; ok {

			return itm.value, ok
		}
	}

	return nil, false
}

// Delete 删除.
func (box *Block) Delete(key string) bool {
	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	box.Lock()
	defer box.Unlock()

	if bucket := box.buckets[off]; bucket != nil {
		if _, ok := bucket[key]; ok {
			delete(bucket, key)

			return true
		}
	}
	delete(box.channels, key)

	return false
}

// GetAndTimeOut 获取值带有超时限制.
func (box *Block) GetAndTimeOut(key string, timeout time.Duration, ch chan interface{}) (b []byte, err error) {
	c := box.registerMessage(key)
	b, ok := box.Get(key)
	if err != nil || ok {

		return b, err
	}

	tick := time.NewTicker(timeout)
	defer tick.Stop()
	select {
	case <-c:
		b, ok = box.Get(key)
	case <-ch:
		err = errors.New("EOF")
	case <-tick.C:
		err = errors.New("timeout")
	}
	box.deregisterMessage(key, c)

	return
}

// registerMessage 获取一个通信对象.
func (box *Block) registerMessage(key string) chan interface{} {
	c := make(chan interface{}, 2)
	box.Lock()
	defer box.Unlock()

	channels, ok := box.channels[key]
	if !ok {
		channels = make(map[chan interface{}]interface{}, 2)
		box.channels[key] = channels
	}
	channels[c] = nil

	return c
}

// deregisterMessage 删除一个通信对象.
func (box *Block) deregisterMessage(key string, c chan interface{}) error {
	box.Lock()
	defer box.Unlock()

	if channels, ok := box.channels[key]; ok {
		delete(channels, c)
		// 空数据删除对象.
		if len(channels) < 1 {
			delete(box.channels, key)
		}
	}

	return nil
}

// StartAndGC 开启垃圾回收.
func (box *Block) StartAndGC() error {
	go box.vaccuum()

	return nil
}

// ClearAll 清空缓存.
func (box *Block) ClearAll() error {
	box.Lock()
	defer box.Unlock()

	for _, channels := range box.channels {
		for channel := range channels {
			channel <- nil
		}
	}
	box.buckets = make([]map[string]*Item, BlockSize)
	box.channels = make(map[string]map[chan interface{}]interface{}, 0)

	return nil
}

// vaccuum 垃圾回收.
func (box *Block) vaccuum() {
	tick := time.Tick(box.dur)
	for {
		<-tick
		if box.buckets == nil {
			continue
		}

		for _, bucket := range box.buckets {
			for key := range bucket {
				box.itemExpired(key)
			}
		}
	}
}

// itemExpired 处理数据是否被回收.
func (box *Block) itemExpired(key string) bool {
	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	box.Lock()
	defer box.Unlock()

	if bucket := box.buckets[off]; bucket != nil {
		if itm, ok := bucket[key]; ok {
			if itm.isExpire() {
				delete(bucket, key)
				if channels, ok := box.channels[key]; ok {
					for channel := range channels {
						channel <- nil
					}
					delete(box.channels, key)
				}

				return true
			}
		}
	}

	return false
}

// isExpire 判定数据是否过期.
func (itm *Item) isExpire() bool {
	if itm.lifespan == 0 {

		return false
	}

	return time.Now().Sub(itm.createTime) > itm.lifespan
}
