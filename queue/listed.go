package queue

// Listed 链表接口.
type Listed interface {
	// 入队.
	Put(string) error
	// 出队.
	Out() (string, bool, error)
	// 队列长度.
	Length() int
}

// head 头部数据结构体.
type head struct {
	len   int   // 链表数据长度.
	first *body // 链表第一个位置.
	last  *body // 链表最后一个位置.
}

// body 数据.
type body struct {
	value string // 数据.
	next  *body  // 下一个数据地址.
}

// Put 添加一个数据.
func (h *head) Put(b string) error {
	if h.last == nil {
		h.last = &body{
			value: b,
			next:  nil,
		}
		h.first = h.last
	} else {
		h.last.next = &body{
			value: b,
			next:  nil,
		}
		h.last = h.last.next
	}
	h.len++

	return nil
}

// 取出一个数据.
func (h *head) Out() (string, bool, error) {
	if h.first == nil {

		return "", false, nil
	}

	val := h.first.value
	h.len--

	if h.first.next != nil {
		h.first = h.first.next
	} else {
		h.first = nil
		h.last = nil
	}

	return val, true, nil
}

// 获取数据长度.
func (h *head) Length() int {

	return h.len
}

// NewListed 新建一个链表.
func NewListed() Listed {

	return &head{
		len:   0,
		first: nil,
		last:  nil,
	}
}
