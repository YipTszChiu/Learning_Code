package geecache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

//测试回调函数
func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	except := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, except) {
		t.Errorf("callback failed")
	}
}

//用一个map模拟耗时的数据库
var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

//测试Get方法
func TestGet(t *testing.T) {
	//loadCounts的map用于记录从db中读取到cache中的次数
	loadCounts := make(map[string]int, len(db))
	//新建一个Group，名为scores，容量为2^10Byte，并实现Getter接口的Get方法
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			//从db中搜索key
			log.Println("[SlowDB] search key", key)
			//如果成功从db搜索到对应的key
			if v, ok := db[key]; ok {
				//接着在loadCounts中检查是否从db往cache读取过该key
				if _, ok := loadCounts[key]; !ok {
					//loadCounts中没有该key的读取记录，初始化一个 "key" - int 映射
					loadCounts[key] = 0
				}
				//如果从loadCounts中已有该key的读取次数记录，次数加1
				loadCounts[key] += 1
				//将从db获取到的value以[]byte格式返回
				return []byte(v), nil
			}
			//如果没有从db搜索到该key，返回空值并报错
			return nil, fmt.Errorf("%s not exsist", key)
		}))

	for k, v := range db {
		//首次调用时Group的Get时cache应该为空，会调用Getter的回调函数从db往cache中加入一个元素
		if view, err := gee.Get(k); err != nil || view.String() != v {
			//如果在gee这个Group中调用Get，发生错误，或者获得的ByteView结果与数据库中实际的v不一致
			t.Fatalf("failed to get value of Tom")
		}
		//再次调用Group的Get，无错误的话缓存命中
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			//此时是直接从cache缓存中获得这个kv，因此loadCount不会再+1
			t.Fatalf("cache %s miss", k)
		}
	}

	//从Group中查询一个不存在的key查看是否命中
	if view, err := gee.Get("unknow"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
