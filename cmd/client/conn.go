package client

import (
	"crypto/tls"
	"fmt"
	"net/url"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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
