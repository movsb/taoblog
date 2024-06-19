package clients

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Client interface {
	proto.UtilsClient
	proto.TaoBlogClient
	proto.ManagementClient
	proto.SearchClient
}

type _Client struct {
	proto.UtilsClient
	proto.TaoBlogClient
	proto.ManagementClient
	proto.SearchClient
}

func NewFromGrpcAddr(addr string) Client {
	cc := NewConn("", addr)
	pc := NewProtoClient(cc, "")
	return &_Client{
		UtilsClient:      pc.Utils,
		TaoBlogClient:    pc.Blog,
		ManagementClient: pc.Management,
		SearchClient:     pc.Search,
	}
}

func NewProtoClient(cc *grpc.ClientConn, token string) *ProtoClient {
	return &ProtoClient{
		cc:         cc,
		token:      token,
		Utils:      proto.NewUtilsClient(cc),
		Blog:       proto.NewTaoBlogClient(cc),
		Management: proto.NewManagementClient(cc),
		Search:     proto.NewSearchClient(cc),
	}
}

type ProtoClient struct {
	cc    *grpc.ClientConn
	token string

	Utils      proto.UtilsClient
	Blog       proto.TaoBlogClient
	Management proto.ManagementClient
	Search     proto.SearchClient
}

func (c *ProtoClient) Context() context.Context {
	return c.contextFrom(context.Background())
}

func (c *ProtoClient) contextFrom(parent context.Context) context.Context {
	if c.token == "" {
		return parent
	}
	return metadata.NewOutgoingContext(parent, metadata.Pairs(`Authorization`, fmt.Sprintf("%s %d:%s", auth.TokenName, auth.AdminID, c.token)))
}

// grpcAddress 可以为空，表示使用 api 同一地址。
// TODO 私有化，并把 token 设置到 default call options
func NewConn(api, grpcAddress string) *grpc.ClientConn {
	secure := false
	if grpcAddress == `` {
		u, _ := url.Parse(api)
		grpcAddress = u.Hostname()
		grpcPort := u.Port()
		if u.Scheme == `http` {
			secure = false
			if grpcPort == `` {
				grpcPort = `80`
			}
		} else {
			secure = true
			if grpcPort == `` {
				grpcPort = `443`
			}
		}

		grpcAddress = fmt.Sprintf(`%s:%s`, grpcAddress, grpcPort)
	}

	options := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100<<20),
			grpc.MaxCallSendMsgSize(100<<20),
		),
		grpc.WithTransportCredentials(utils.IIF(secure, credentials.NewTLS(&tls.Config{}), insecure.NewCredentials())),
	}

	conn, err := grpc.Dial(grpcAddress, options...)
	if err != nil {
		panic(err)
	}

	return conn
}
