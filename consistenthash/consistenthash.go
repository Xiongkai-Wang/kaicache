package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 分布式缓存 --> multiple nodes
//	if key does not exist in current node, how to find in other nodes?
// 		iterate all nodes? ===> time consuming
// 		choose one node?   ====> with hash algorithm: (hash(key) % nodeNum) ensure one key everytime points to the node
//  Problem: what if nodeNum changes(one node crashed)?
// 			then (缓存雪崩)几乎缓存值对应的节点都发生了改变,即几乎所有的缓存值都失效了,造成瞬时DB请求量大、压力骤增，引起雪崩

// then we use Consistent hash
type Hash func(data []byte) uint32

// 定义了函数类型Hash，采取依赖注入的方式，允许用于替换成自定义的Hash函数，也方便测试时替换，默认为crc32.ChecksumIEEE

type Map struct {
	hash     Hash
	replicas int            // one real node maps to N virtual nodes
	keys     []int          // sorted   hash circle
	hashMap  map[int]string // vitual nodes map to real nodes
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// add Nodes
func (m *Map) Add(keys ...string) { // can add 1 & more at a time
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// strconv.Itoa(i) : FormatInt(int64(i), 10).
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key // virtual node ==> real node
		}
	}
	sort.Ints(m.keys) // sort a slice in increasing order
}

// Get gets the closest node in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	// Binary search 顺时针找到第一个匹配的*虚拟节点*的下标 idx
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	return m.hashMap[m.keys[idx%len(m.keys)]]
	// idx%len(m.keys) 是因为 sort.Search 是在[0,len(m.keys)]的范围
}
