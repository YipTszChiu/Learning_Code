package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser // conn 是由构建函数传入，通常是通过 TCP 或者 Unix 建立 socket 时得到的链接实例
	buf  *bufio.Writer      // 防止阻塞创建的带缓冲的 Writer
	dec  *gob.Decoder       // 对应 gob 的 Decoder
	enc  *gob.Encoder       // 对应 gob 的 Encoder
}

// 检测GobCodec是否实现了Codec接口；使用 _ 是避免编译时有未使用的变量；如果没有实现接口，编译时会报错
var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}

// 使用gob进行解码
func (c *GobCodec) ReadHeader(h *Header) error {
	return c.dec.Decode(h)
}

// 使用gob进行解码
func (c *GobCodec) ReadBody(body interface{}) error {
	return c.dec.Decode(body)
}

// 使用gob对header和body进行编码
func (c *GobCodec) Write(h *Header, body interface{}) (err error) {
	// 延迟执行：
	defer func() {
		// Flush 方法使得缓存都写入 io.Writer 对象中
		_ = c.buf.Flush()
		// 如果对header或body编码发生错误，则调用自实现的Close()方法关闭GobCodec的io链接实例
		if err != nil {
			_ = c.Close()
		}
	}()
	// 使用 gob 对 header 进行编码，如果有错误，记录日志并返回错误信息
	if err := c.enc.Encode(h); err != nil {
		log.Println("rpc codec: gob error encoding header:", err)
		return err
	}
	// 使用 gob 对 body 进行编码，如果有错误，记录日志并返回错误信息
	if err := c.enc.Encode(body); err != nil {
		log.Println("rpc codec: gob error encoding body:", err)
		return err
	}
	return nil
}

func (c *GobCodec) Close() error {
	return c.conn.Close()
}
