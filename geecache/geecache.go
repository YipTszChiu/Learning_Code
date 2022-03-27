package geecache

import (
	"fmt"
	"log"
	"sync"
)

//回调Getter
//Getter接口，通过key加载数据
type Getter interface {
	Get(key string) ([]byte, error)
}

//GetterFunc实现Getter接口——接口型函数
type GetterFunc func(key string) ([]byte, error)

//Get实现Getter的接口函数
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//核心数据结构Group
//一个Group可以认为是缓存的命名空间，每个Group有唯一的name。比如可以有成绩sources，学生信息info
type Group struct {
	name      string
	getter    Getter //缓存未命中时的回调callback
	mainCache cache  //前面实现的并发缓存
}

var (
	mu     sync.RWMutex //一写多读的互斥锁
	groups = make(map[string]*Group)
)

//NewGroup用于新建一个Group的实例
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	//传入空Getter处理
	if getter == nil {
		panic("nil Getter")
	}

	//加锁并延迟解锁
	mu.Lock()
	defer mu.Unlock()

	//新建一个Group
	g := &Group{
		name,
		getter,
		cache{cacheBytes: cacheBytes},
	}
	//将这个Group加入到map映射中
	groups[name] = g

	return g
}

//GetGroup返回名为name的Group，如果没有对应的Group则返回nil
func GetGroup(name string) *Group {
	mu.RLock() //注意是ReadLock只读锁， 不涉及写操作
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	//空key处理
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	//如果缓存命中，写日志
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}
	//如果缓存未命中，需要加载
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	//调用Getter接口的Get方法
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	//value为返回信息的副本
	value := ByteView{cloneBytes(bytes)}
	//调用pupulateCache调整cache
	g.populateCache(key, value)

	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	//调用cache封装的add方法，包含了移动链表元素、检测是否超过容量等操作
	g.mainCache.add(key, value)
}
