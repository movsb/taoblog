package proto

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	"github.com/movsb/taoblog/modules/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func NewProtoClient(cc *grpc.ClientConn, token string) *ProtoClient {
	return &ProtoClient{
		cc:         cc,
		token:      token,
		Blog:       NewTaoBlogClient(cc),
		Management: NewManagementClient(cc),
	}
}

type ProtoClient struct {
	cc    *grpc.ClientConn
	token string

	Blog       TaoBlogClient
	Management ManagementClient
}

func (c *ProtoClient) Context() context.Context {
	return c.ContextFrom(context.Background())
}

func (c *ProtoClient) ContextFrom(parent context.Context) context.Context {
	return metadata.NewOutgoingContext(parent, metadata.Pairs(auth.TokenName, c.token))
}

// grpcAddress 可以为空，表示使用 api 同一地址。
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

	var conn *grpc.ClientConn
	var err error
	if secure {
		if conn, err = grpc.Dial(
			grpcAddress,
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100<<20)),
		); err != nil {
			panic(err)
		}
	} else {
		if conn, err = grpc.Dial(
			grpcAddress,
			grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100<<20)),
		); err != nil {
			panic(err)
		}
	}

	return conn
}
