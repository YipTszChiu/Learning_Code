package lru

import "container/list"

//参考：https://geektutu.com/post/geecache-day1.html

//LRU cache,对并发访问不安全
type Cache struct {
	maxBytes int64      //最大容量
	nBytes   int64      //已用容量
	ll       *list.List //实际存储条目的双向链表
	//字典，使从key查找到value的时间复杂度O(1)，往字典插入一条记录的时间复杂度也为O(1)
	cache map[string]*list.Element
	//可选属性，在删除条目时执行的回调函数
	OnEvicted func(key string, value Value)
}

//双向链表中所存的条目，字典中有了kv映射仍要在链表中存key的原因是：淘汰节点时需要用key从字典中删除对应的映射
type entry struct {
	key   string
	value Value
}

//为了增加通用性，所存的Value可以为任意实现了Len()方法的类型
type Value interface {
	Len() int
}

//用于实例化的New()函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//查找功能
func (c *Cache) Get(key string) (value Value, ok bool) {
	//查找字典cache里是否存在对应的kv
	if elem, ok := c.cache[key]; ok {
		//如果查找成功，将该节点移动到Front
		c.ll.MoveToFront(elem)
		//elem是一个*List.Element，将其强转为*entry
		kv := elem.Value.(*entry)
		return kv.value, true
	}
	//查找失败直接返回默认值
	return
}

//缓存淘汰
func (c *Cache) Removeoldest() {
	//获取最近最少使用的条目
	elem := c.ll.Back()
	//判空
	if elem != nil {
		//从链表中删除该条目
		c.ll.Remove(elem)
		//强转为条目类型
		kv := elem.Value.(*entry)
		//从字典中删除映射
		delete(c.cache, kv.key)
		//已用容量减少删除掉的条目大小
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len()) //value可以为任何格式，因此需要调用Len()
		//如果Cache有定义删除回调函数，需要返回相应的值
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

//新增/修改
func (c *Cache) Add(key string, value Value) {
	//如果字典中已存在该key，执行修改
	if elem, ok := c.cache[key]; ok {
		//将该节点移动到Front
		c.ll.MoveToFront(elem)
		//强转类型
		kv := elem.Value.(*entry)
		//更新已用内存
		c.nBytes += int64(kv.value.Len()) - int64(value.Len())
		//由于kv是*entry类型，因此可以修改底层的value
		kv.value = value
	} else { //字典中不存在该key，执行新增
		//直接将新建的条目加入到Front
		elem := c.ll.PushFront(&entry{key, value})
		//将新节点与字典映射
		c.cache[key] = elem
		//更新已用内存
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	//如果超过Cache最大容量，需要删除条目
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.Removeoldest()
	}
}

//为了方便测试，重写Cache的Len()，使其返回已经插入的条目数量
func (c *Cache) Len() int {
	return c.ll.Len()
}
