package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash函数将[]byte字符切片映射为32位的int
type Hash func(data []byte) uint32

//	Map包含了所有散列的键
type Map struct {
	hash     Hash           // 哈希函数
	replicas int            // 虚拟节点倍数
	keys     []int          // Sorted  哈希环
	hashMap  map[int]string // 虚拟节点和真实节点的映射表
}

// 实例化Map，采用依赖注入，允许自定义虚拟节点倍数，允许自定义Hash函数，默认为crc32.ChecksumIEEE算法
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// 添加节点/机器的 Add() 方法
// 传入 0 或 多个 真实节点的名称以添加节点
func (m *Map) Add(keys ...string) {
	// 遍历每一个添加的真是节点
	for _, key := range keys {
		// 每个真实节点都创建 m.replicas 个虚拟节点
		for i := 0; i < m.replicas; i++ {
			// 虚拟节点的名称为 i + key
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 将虚拟节点名称哈希后加到环上
			m.keys = append(m.keys, hash)
			// 在hashMap上增加虚拟节点与真实节点的映射
			m.hashMap[hash] = key
		}
	}
	// 将环上的哈希值排序
	sort.Ints(m.keys)
}

// 实现选择节点的 Get()方法
// 获取hash环上最近的节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	// 将key哈希
	hash := int(m.hash([]byte(key)))
	// 二分查找对应的虚拟节点
	// 第一个参数是范围，第二个参数是自定义函数，规则是找到第一个表达式为true的index
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 环装结构需要取余，通过hashMap返回真实节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
