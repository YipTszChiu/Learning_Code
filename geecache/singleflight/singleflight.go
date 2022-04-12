package singleflight

import "sync"

// call代表进行中或已结束的请求，waitgroup锁避免重入
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// singleflight的主体结构，管理不同key的请求(call)
type Group struct {
	mu sync.Mutex       // mu用于保护成员变量 m 不被并发读写
	m  map[string]*call // m 维护 key 与其对应的 call 请求
}

// 对相应的key请求进行处理，传入匿名函数fn用于获取key对应的val，Do函数使用sync.WaitGruop使得同一个key无论调用多少次Do，fn只会执行一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 加锁保护map
	g.mu.Lock()
	// 延迟初始化
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 如果已经有关于该key的请求，Wait等待对应key的所有请求结束再返回
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err // TODO：val如何获得？
	}
	// 如果该key目前没有请求，新建一个并使WaitGroup锁的计数+1
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c

	// call执行传进来的fn函数获取val
	c.val, c.err = fn()
	// Done使WaitGroup锁的计数-1
	c.wg.Done()

	// call请求结束后，对map进行加锁，并删除该key对应的映射
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
