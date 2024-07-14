package dialers

import (
	"bytes"
	"encoding/gob"
	"io"
	"log"
	"net"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/xtaci/smux"
)

type ClientOrServerStream interface {
	Send(*proto.DialRemoteMessage) error
	Recv() (*proto.DialRemoteMessage, error)
}
type StreamAsConn struct {
	s ClientOrServerStream
	b *bytes.Buffer
}

// 用于把 GRPC STREAM 连接转换成连续的 读/写 接口。
func NewStreamAsConn(s ClientOrServerStream) io.ReadWriteCloser {
	return &StreamAsConn{
		s: s,
		b: bytes.NewBuffer(nil),
	}
}

func (c *StreamAsConn) Read(p []byte) (int, error) {
	if c.b.Len() > 0 {
		return c.b.Read(p)
	}
	m, err := c.s.Recv()
	if err != nil {
		return 0, err
	}
	c.b.Write(m.Data)
	return c.b.Read(p)
}

// 只有 client 需要关，那就自行关。
func (c *StreamAsConn) Close() error {
	return nil
}

func (c *StreamAsConn) Write(p []byte) (int, error) {
	m := proto.DialRemoteMessage{
		Data: p,
	}
	if err := c.s.Send(&m); err != nil {
		return 0, err
	}
	return len(p), nil
}

type RemoteDialerManager struct {
	mux *smux.Session
}

func NewRemoteDialerManager(s ClientOrServerStream) *RemoteDialerManager {
	return &RemoteDialerManager{
		mux: utils.Must1(smux.Client(NewStreamAsConn(s), nil)),
	}
}

func (m *RemoteDialerManager) Run() error {
	<-m.mux.CloseChan()
	return nil
}

type DialRemoteRequest struct {
	Addr string
}

type DialRemoteResponse struct {
	Error string
}

func (m *RemoteDialerManager) Dial(addr string) (net.Conn, error) {
	conn, err := m.mux.OpenStream()
	if err != nil {
		return nil, err
	}

	enc := gob.NewEncoder(conn)
	if err := enc.Encode(DialRemoteRequest{
		Addr: addr,
	}); err != nil {
		conn.Close()
		return nil, err
	}

	var resp DialRemoteResponse
	if err := gob.NewDecoder(conn).Decode(&resp); err != nil {
		conn.Close()
		return nil, err
	}

	if resp.Error != "" {
		conn.Close()
		log.Println(addr, err)
		return nil, err
	}

	return conn, nil
}
