package codec

import "io"

// 请求和响应中除了参数和返回值的剩余信息
type Header struct {
	ServiceMethod string // 格式 "Service.Method" 服务名.方法名
	Seq           uint64 // 请求的序号，用于区分不同的请求
	Error         string //错误信息，客户端置为空，如果服务端发生错误，将错误信息置于Error中
}

// 抽象出对消息体进行编解码的接口Codec，为了实现不同的Codec实例
type Codec interface {
	io.Closer
	ReadHeader(*Header) error
	ReadBody(interface{}) error
	Write(*Header, interface{}) error
}

// 抽象出Codec的构造函数，客户端和服务端可以通过 Codec 的 Type 得到构造函数，从而创建 Codec 实例。
// 这部分代码和工厂模式类似，与工厂模式不同的是，返回的是构造函数，而非实例
type NewCodecFunc func(io.ReadWriteCloser) Codec

type Type string

// 定义两种Codec，在实际代码中只实现 Gob 一种
const (
	GobType  Type = "application/gob"
	JsonType Type = "application/json" // not implemented
)

var NewCodecFuncMap map[Type]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[Type]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
