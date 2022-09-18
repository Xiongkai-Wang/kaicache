package lru

import "container/list"

// cache is a LRU cache
type Cache struct {
	maxBytes  int64                         // 允许使用的最大内存
	nbytes    int64                         // 当前已使用的内存
	ll        *list.List                    //  Go 语言标准库实现的双向链表list.List,
	cache     map[string]*list.Element      // 键是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // OnEvicted 是某条记录被移除时的回调函数，可以为 nil
}

// 键值对 entry 是双向链表节点的数据类型，
// 在链表中仍保存每个值对应的 key 的好处在于，淘汰队首节点时，需要用key从字典中删除对应的映射
type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes
// 为了通用性，我们允许值是实现了 Value 接口的任意类型
type Value interface {
	Len() int
}

// constrcutor
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 查找功能
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)    // 算法核心：将最近被访问的数据放置到队尾
		kv := ele.Value.(*entry) // list 存储的是任意类型interface，使用时需要类型转换
		return kv.value, true
	}
	return
}

// 淘汰功能
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back() // 取到队首节点，从链表中删除
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len()) // len() built-in func; Len() self-defined
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 新增功能
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// key exists: update it
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// do not exist: add
		ele := c.ll.PushFront(&entry{key: key, value: value}) // PushFront returns *Element
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// for 是没问题的，可能会 remove 多次，
	// 添加一条大的键值对，可能需要淘汰掉多个键值对，直到 nbytes < maxBytes。
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}
