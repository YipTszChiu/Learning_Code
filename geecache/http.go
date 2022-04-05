package geecache

import (
	"Learning_Code/geecache/consistenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

// 服务端

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self        string                 // 记录自己的地址，包括主机名/IP和端口
	basePath    string                 // 节点间通讯地址的前缀，默认是/_geecache/
	mu          sync.Mutex             // 保护peers和httpGetters
	peers       *consistenthash.Map    // 根据具体的key选择节点
	httpGetters map[string]*httpGetter // 映射远程节点与对应的httpGetter e.g. "http://10.0.0.2:8008"

	//那么 http://example.com/_geecache/ 开头的请求，就用于节点间的访问。
	//因为一个主机上还可能承载其他的服务，加一段 Path 是一个好习惯。比如，大部分网站的 API 接口，一般以 /api 作为前缀
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
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

// 实例化一致性哈希，并添加节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 初始化一个一致性哈希的Map，并调用Add函数增加节点
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	// 建立每个peer与httpGetter的映射
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// 实现PeerPicker接口，通过key选择对应的peer，返回节点对应的 HTTP 客户端。
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// 通过一致性哈希环找到应该读取的节点
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}

	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

// 客户端

type httpGetter struct {
	baseURL string //baseURL 表示将要访问的远程节点的地址，例如 http://example.com/_geecache/
}

// 实现PeerGetter接口的Get方法
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	// 使用 http.Get() 方式获取返回值，并转换为 []bytes 类型。
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil)
