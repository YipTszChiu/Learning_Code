package geecache

import (
	"Learning_Code/geecache/lru"
	"sync"
)

//为lru.Cache增加并发特性

type cache struct {
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	//加锁，延迟解锁
	c.mu.Lock()
	defer c.mu.Unlock()

	//延迟初始化：该对象的创建将会延迟至第一次使用该对象时。主要用于提高性能，并减少程序内存要求。
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}

	//调用Add添加
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	//加锁，延迟解锁
	c.mu.Lock()
	defer c.mu.Unlock()

	//空指针处理
	if c.lru == nil {
		return
	}

	//不为空则调用Get查找
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
