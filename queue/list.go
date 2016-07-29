package queue

import (
	"errors"
	"sync"
	"time"

	"../h32"
)

// queue 队列结构体.
type queue struct {
	sync.RWMutex                                // 读写锁.
	tube map[string]*li                         // 队列排队.
	log  map[interface{}]map[string]interface{} // 记录单个任务doing状态的key.
	db   []map[string]*job                      // 原始数据.
	dur  time.Duration                          // 垃圾回收周期.
}

// job 任务信息.
type job struct {
	tube   string // 消息队列名称.
	key    string // 唯一ID标示.
	value  []byte // 数据.
	status uint8  // 任务状态.
}

// li 任务连.
type li struct {
	list       Listed                           // 链表.
	channels   map[chan interface{}]interface{} // 消息订阅.
	updateTime time.Time                        // 更新时间.
}

// Join 向队列中，添加一个任务.
func (Q *queue) Join(tube, key string, value []byte) error {
	_, ok := Q.GetDb(key)
	if ok {

		return nil
	}
	Q.Lock()
	defer Q.Unlock()

	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	bucket := Q.db[off]
	if bucket == nil {
		bucket = make(map[string]*job, 1)
		Q.db[off] = bucket
	}
	bucket[key] = &job{
		tube:   tube,
		key:    key,
		value:  value,
		status: READY,
	}

	tubes, ok := Q.tube[tube]
	if !ok {
		tubes = &li{
			list:     NewListed(),
			channels: make(map[chan interface{}]interface{}, 1),
		}
		Q.tube[tube] = tubes
	} else {
		for channel := range tubes.channels {
			channel <- nil
		}
		tubes.channels = make(map[chan interface{}]interface{}, 1)
	}
	tubes.list.Put(key)
	tubes.updateTime = time.Now()

	return nil
}

// Finish 完成一个任务.
func (Q *queue) Finish(key string, conn interface{}) (bool, error) {
	Q.Lock()
	defer Q.Unlock()

	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	if bucket := Q.db[off]; bucket != nil {
		if itm, ok := bucket[key]; ok {
			itm.status = DELAYED
			if logs, ok := Q.log[conn]; ok {
				delete(logs, key)
			}

			// 删除log中的数据.
			if m, ok := Q.log[conn]; ok {
				delete(m, key)
			}
			return true, nil
		}
	}

	return false, nil
}

// GetAndDoing 获取一个任务，修改任务状态为正在开始中.
func (Q *queue) GetAndDoing(tube string, conn interface{}) (key string, value []byte, err error) {
	Q.Lock()
	defer Q.Unlock()

	if tubes, ok := Q.tube[tube]; ok {
		GO:
		if key, ok, err := tubes.list.Out(); ok && err == nil {
			off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
			if bucket := Q.db[off]; bucket != nil {
				if itm, ok := bucket[key]; ok {
					if itm.status == READY {
						itm.status = RESERVED
						logs, ok := Q.log[conn]
						if !ok {
							logs = make(map[string]interface{}, 1)
							Q.log[conn] = logs
						}
						logs[key] = nil

						return key, itm.value, err
					}
					goto GO
				}
			}
		}
	}

	return "", nil, err
}

// Exists 判定一个人是否存在, 该任务必须为未开始，正在完成中.
func (Q *queue) Exists(key string) bool {
	Q.Lock()
	defer Q.Unlock()

	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	if bucket := Q.db[off]; bucket != nil {
		if itm, ok := bucket[key]; ok {
			if itm.status == READY || itm.status == RESERVED {

				return true
			}
		}
	}

	return false
}

// StartAndGC 启动垃圾回收.
func (Q *queue) StartAndGC() error {
	go Q.vaccuum()

	return nil
}

// vaccuum 周期性检查数据.
func (Q *queue) vaccuum() {
	tick := time.Tick(Q.dur)

	for {
		<-tick
		if Q.db == nil {
			continue
		}

		for _, jobs := range Q.db {
			for key := range jobs {
				Q.itemExpired(key)
			}
		}

		for name := range Q.tube {
			Q.itemExpiredQueue(name)
		}
	}
}

// itemExpiredQueue 清除过时队列.
func (Q *queue) itemExpiredQueue(queue string) error {
	Q.Lock()
	defer Q.Unlock()

	if list, ok := Q.tube[queue]; ok {
		if time.Now().Sub(list.updateTime) > time.Hour * 24 {
			delete(Q.tube, queue)
		}
	}

	return nil
}

// itemExpired 判定队列是否及完成，如果已经已经完成，则删除.
func (Q *queue) itemExpired(key string) {
	Q.Lock()
	defer Q.Unlock()

	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	if bucket := Q.db[off]; bucket != nil {
		if job, ok := bucket[key]; ok {
			if job.status == DELAYED {
				delete(bucket, key)
			}
		}
	}
}

// RestoreOne 还原一个任务.
func (Q *queue) RestoreOne(key string, conn interface{}) (bool, error) {
	Q.Lock()
	defer Q.Unlock()

	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	if bucket := Q.db[off]; bucket != nil {
		if itm, ok := bucket[key]; ok {
			if itm.status == RESERVED {
				itm.status = READY
				tubes, ok := Q.tube[itm.tube]
				if !ok {
					tubes = &li{
						list:     NewListed(),
						channels: make(map[chan interface{}]interface{}, 1),
					}
					Q.tube[itm.tube] = tubes
				} else {
					for channel := range tubes.channels {
						channel <- nil
					}
					tubes.channels = make(map[chan interface{}]interface{}, 1)
				}
				tubes.list.Put(key)
				if logs, ok := Q.log[conn]; ok {
					delete(logs, key)
				}

				return true, nil
			}
		}
	}

	return false, nil
}

// Usr1 添加一个通知，如果有数据，通知用户.
func (Q *queue) Usr1(tube string, ch chan interface{}) (ok bool, err error) {
	ok, err = Q.existsQueue(tube)
	if err != nil || ok {

		return
	}

	c := Q.registerMessage(tube)
	ok, err = Q.existsQueue(tube)
	if err != nil || ok {
		Q.deregisterMessage(tube, c)

		return
	}

	select {
	case <-c:
		ok, err = Q.existsQueue(tube)
	case <-ch:
		err = errors.New("EOF")
	}
	Q.deregisterMessage(tube, c)

	return ok, err
}

// registerMessage 设置一个消息通知.
func (Q *queue) registerMessage(tube string) chan interface{} {
	c := make(chan interface{}, 2)
	Q.Lock()
	defer Q.Unlock()

	tubes, ok := Q.tube[tube]
	if !ok {
		tubes = &li{
			list:     NewListed(),
			channels: make(map[chan interface{}]interface{}),
		}
		Q.tube[tube] = tubes
	}
	tubes.channels[c] = nil

	return c
}

// existsQueue 判定指定消息队列中，是否存在队列.
func (Q *queue) existsQueue(tube string) (bool, error) {
	Q.RLock()
	defer Q.RUnlock()

	if tubes := Q.tube[tube]; tubes != nil {
		if tubes.list.Length() > 0 {

			return true, nil
		}
	}

	return false, nil
}

// deregisterMessage  注销一个消息.
func (Q *queue) deregisterMessage(tube string, c chan interface{}) error {
	Q.Lock()
	defer Q.Unlock()

	if tubes, ok := Q.tube[tube]; ok {
		delete(tubes.channels, c)

		return nil
	}

	return nil
}

// GetDb 获取DB数据.
func (Q *queue) GetDb(key string) ([]byte, bool) {
	Q.RLock()
	defer Q.RUnlock()

	off := h32.DefaultHash.GetOffset(key, BlockSize, BucketSize)
	if bucket := Q.db[off]; bucket != nil {
		if itm, ok := bucket[key]; ok {

			return itm.value, ok
		}
	}

	return nil, false
}

// RestoreAll 还原一个连接对象正在做的任务进行还原.
func (Q *queue) RestoreAll(conn interface{}) error {
	if logs, ok := Q.log[conn]; ok && logs != nil {
		for key := range logs {
			Q.RestoreOne(key, conn)
		}

		Q.Lock()
		defer Q.Unlock()
		delete(Q.log, conn)
	}

	return nil
}
