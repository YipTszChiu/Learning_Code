package main

import "Learning_Code/mysql_learning"

// geecache测试
//var db = map[string]string{
//	"Tom":  "630",
//	"Jack": "589",
//	"Sam":  "567",
//}

// 测试http.go
//func main() {
//	//新建一个名为"scores"的group，并定义get方法，当缓存为空时从db获取对应的数据
//	geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
//		func(key string) ([]byte, error) {
//			log.Println("[SlowDB] search key", key)
//			if v, ok := db[key]; ok {
//				return []byte(v), nil
//			}
//			return nil, fmt.Errorf("%s not exist", key)
//		}))
//
//	addr := "localhost:9999"
//	peers := geecache.NewHTTPPool(addr)
//	log.Println("geecache is running at", addr)
//	//使用 http.ListenAndServe 在 9999 端口启动了 HTTP 服务。
//	log.Fatal(http.ListenAndServe(addr, peers))
//}
//

// mysql测试
func main() {
	mysql_learning.Main()
}
