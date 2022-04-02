package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache/"

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self     string //记录自己的地址，包括主机名/IP和端口
	basePath string //节点间通讯地址的前缀，默认是/_geecache/

	//那么 http://example.com/_geecache/ 开头的请求，就用于节点间的访问。
	//因为一个主机上还可能承载其他的服务，加一段 Path 是一个好习惯。比如，大部分网站的 API 接口，一般以 /api 作为前缀
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self,
		defaultBasePath,
	}
}

// Log函数用于按格式打印日志
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

//实现ServeHTTP方法，任何实现该方法的对象都可以作为HTTP的Handler
//log info with server name
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//判断访问路径是否前缀是否为 basePath
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	// 记录日志
	p.Log("%s %s", r.Method, r.URL.Path)

	// 约定访问路径格式为 /<basepath>/<groupname>/<key>

	// 将访问路径舍弃掉basePath部分，再按 "/" 分割获得 groupname 和 key
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	//如果分割后[]string数组元素不为2则报错
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	// 获取分割成功的 groupname 以及 key
	groupName := parts[0]
	key := parts[1]

	//调用GetGroup获取对应的group
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	//调用group实现的Get方法获取已经缓存的kv
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	//使用 w.Write() 将缓存值作为 httpResponse 的 body 返回
	w.Write(view.ByteSlice())
}
