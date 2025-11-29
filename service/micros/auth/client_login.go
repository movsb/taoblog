package micros_auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/phuslu/lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type _ClientLoginSessionData struct {
	token string
}

type ClientLoginService struct {
	// 映射挑战数据到会话数据。
	clientLoginSessions *lru.TTLCache[string, *_ClientLoginSessionData]

	getHome func() string

	proto.UnimplementedClientLoginServer
}

func NewClientLoginService(ctx context.Context, sr grpc.ServiceRegistrar, getHome func() string) *ClientLoginService {
	s := &ClientLoginService{
		clientLoginSessions: lru.NewTTLCache[string, *_ClientLoginSessionData](8),
		getHome:             getHome,
	}

	proto.RegisterClientLoginServer(sr, s)

	return s
}

func (s *ClientLoginService) SetClientLoginToken(random, token string) {
	s.clientLoginSessions.Set(random, &_ClientLoginSessionData{
		token: token,
	}, time.Second*5)
}

func (s *ClientLoginService) ClientLogin(in *proto.ClientLoginRequest, srv proto.ClientLogin_ClientLoginServer) error {
	var random [16]byte
	rand.Read(random[:])

	u := utils.Must1(url.Parse(s.getHome())).JoinPath(`admin`, `login`, `client`)
	q := u.Query()

	randomString := fmt.Sprintf(`%x`, random)
	q.Set(`random`, randomString)
	u.RawQuery = q.Encode()

	s.clientLoginSessions.Set(randomString, &_ClientLoginSessionData{}, time.Minute*5)

	if err := srv.Send(&proto.ClientLoginResponse{
		Response: &proto.ClientLoginResponse_Open_{
			Open: &proto.ClientLoginResponse_Open{
				AuthUrl: u.String(),
			},
		},
	}); err != nil {
		return err
	}

	for {
		value, found := s.clientLoginSessions.Get(randomString)
		if !found {
			return status.Error(codes.Aborted, `超时已取消授权。`)
		}
		if value.token != `` {
			s.clientLoginSessions.Delete(randomString)
			srv.Send(&proto.ClientLoginResponse{
				Response: &proto.ClientLoginResponse_Success_{
					Success: &proto.ClientLoginResponse_Success{
						Token: value.token,
					},
				},
			})
			return nil
		}
		log.Println(`等待登录授权：`, randomString)
		time.Sleep(time.Second)
	}
}
