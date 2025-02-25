package clients

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/movsb/http2tcp"
	grpc_proxy "github.com/movsb/taoblog/gateway/handlers/grpc"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func NewProtoClient(home string, token string) *ProtoClient {
	cc := _NewConn(home, ``)
	return _NewFromCC(cc, token)
}

func NewProtoClientFromAddress(grpcAddress string) *ProtoClient {
	cc := _NewConn(``, grpcAddress)
	return _NewFromCC(cc, ``)
}

func _NewFromCC(cc *grpc.ClientConn, token string) *ProtoClient {
	return &ProtoClient{
		cc:         cc,
		token:      token,
		Auth:       proto.NewAuthClient(cc),
		Utils:      proto.NewUtilsClient(cc),
		Blog:       proto.NewTaoBlogClient(cc),
		Management: proto.NewManagementClient(cc),
		Search:     proto.NewSearchClient(cc),
		Notify:     proto.NewNotifyClient(cc),
	}
}

type ProtoClient struct {
	cc    *grpc.ClientConn
	token string

	Auth       proto.AuthClient
	Utils      proto.UtilsClient
	Blog       proto.TaoBlogClient
	Management proto.ManagementClient
	Search     proto.SearchClient
	Notify     proto.NotifyClient
}

func (c *ProtoClient) Context() context.Context {
	return c.contextFrom(context.Background())
}

// TODO 配置文件中写/传完整的 ID:TOKEN 格式，而不是假定是 AdminID。
func (c *ProtoClient) contextFrom(parent context.Context) context.Context {
	if c.token == "" {
		return parent
	}
	return metadata.NewOutgoingContext(parent, metadata.Pairs(`Authorization`, fmt.Sprintf("%s %d:%s", auth.TokenName, auth.AdminID, c.token)))
}

// grpcAddress 可以为空，表示使用 api 同一地址。
// TODO 私有化，并把 token 设置到 default call options
func _NewConn(home, orGrpcAddress string) *grpc.ClientConn {
	options := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100<<20),
			grpc.MaxCallSendMsgSize(100<<20),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	var (
		conn *grpc.ClientConn
		err  error
	)

	if home != `` {
		u := utils.Must1(url.Parse(home))
		options = append(options,
			grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
				conn, err := http2tcp.Dial(u.JoinPath(`/v3/grpc`).String(), grpc_proxy.PreSharedKey, ``)
				if err != nil {
					// log.Println(home, err)
					return nil, err
				}
				return &_NetConn{conn}, nil
			}),
		)
		conn, err = grpc.Dial(`does-not-matter:0`, options...)
	} else {
		conn, err = grpc.Dial(orGrpcAddress, options...)
	}

	if err != nil {
		panic(err)
	}

	return conn
}

type _NetConn struct {
	io.ReadWriteCloser
}

func (_NetConn) LocalAddr() net.Addr                { return nil }
func (_NetConn) RemoteAddr() net.Addr               { return nil }
func (_NetConn) SetDeadline(t time.Time) error      { return nil }
func (_NetConn) SetReadDeadline(t time.Time) error  { return nil }
func (_NetConn) SetWriteDeadline(t time.Time) error { return nil }

var _ net.Conn = (*_NetConn)(nil)

func (c *ProtoClient) SendInstant(title, message string) {
	if _, err := c.Notify.SendInstant(
		c.Context(),
		&proto.SendInstantRequest{
			Subject: title,
			Body:    message,
		},
	); err != nil {
		log.Println(err)
		log.Println(title, message)
	}
}
