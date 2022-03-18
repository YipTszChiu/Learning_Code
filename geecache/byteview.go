package geecache

//抽象一个只读数据结构表示缓存值
type ByteView struct {
	b []byte
}

//实现ByteView的Len()方法
func (v ByteView) Len() int {
	return len(v.b)
}

//ByteSlice方法返回一个ByteView的副本，防止被篡改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

//String方法返回字符串格式的ByteView
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
