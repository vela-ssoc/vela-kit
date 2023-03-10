package tunnel

import (
	"github.com/gorilla/websocket"
	"io"
	"sync"
)

// Conn 主连接通道
type Conn struct {
	conn  *websocket.Conn // 底层 websocket 通道
	wmu   sync.Mutex      // write 锁
	ident Ident           // 认证信息
	claim Claim           // 授权信息
}

// Send 发送消息
func (c *Conn) Send(msg *Message) error {
	if msg == nil {
		return nil
	}
	data, err := msg.marshal(c.claim.Mask)
	if err != nil {
		return err
	}

	c.wmu.Lock()
	defer c.wmu.Unlock()

	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Conn) Ident() Ident {
	return c.ident
}

func (c *Conn) Claim() Claim {
	return c.claim
}

func (c *Conn) close() error {
	return c.conn.Close()
}

func (c *Conn) receive() (*Receive, error) {
	_, raw, err := c.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	rec := &Receive{}
	if err = rec.unmarshal(raw, c.claim.Mask); err != nil {
		return nil, err
	}
	return rec, nil
}

type StreamConn struct {
	wMutex sync.Mutex
	reader io.Reader
	conn   *websocket.Conn
}

func newStreamConn(conn *websocket.Conn) *StreamConn {
	return &StreamConn{
		reader: websocket.JoinMessages(conn, ""),
		conn:   conn,
	}
}

func (s *StreamConn) Read(p []byte) (int, error) {
	return s.reader.Read(p)
}

func (s *StreamConn) Write(p []byte) (int, error) {
	s.wMutex.Lock()
	defer s.wMutex.Unlock()

	if e := s.conn.WriteMessage(websocket.BinaryMessage, p); e != nil {
		return 0, e
	}

	return len(p), nil
}

func (s *StreamConn) Close() error {
	return s.conn.Close()
}
