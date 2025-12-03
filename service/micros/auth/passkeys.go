package micros_auth

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/auth/user"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	auth_webauthn "github.com/movsb/taoblog/service/micros/auth/webauthn"
	"github.com/phuslu/lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthBackend interface {
	GetUserByID(id int) (*user.User, error)
}

type PasskeysService struct {
	authBackend AuthBackend
	wa          func() *webauthn.WebAuthn
	cookieGen   func(user *user.User, agent string) []*proto.FinishPasskeysLoginResponse_Cookie

	loginSessions *lru.TTLCache[string, *webauthn.SessionData]

	proto.UnimplementedPasskeysServer
}

func NewPasskeysService(
	ctx context.Context,
	sr grpc.ServiceRegistrar,
	wa func() *webauthn.WebAuthn,
	cookieGen func(user *user.User, agent string) []*proto.FinishPasskeysLoginResponse_Cookie,
) *PasskeysService {
	return &PasskeysService{
		wa:            wa,
		cookieGen:     cookieGen,
		loginSessions: lru.NewTTLCache[string, *webauthn.SessionData](8),
	}
}

// BeginPasskeysLogin implements proto.AuthServer.
func (p *PasskeysService) BeginPasskeysLogin(context.Context, *proto.BeginPasskeysLoginRequest) (*proto.BeginPasskeysLoginResponse, error) {
	options, session, err := p.wa().BeginDiscoverableLogin()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	p.loginSessions.Set(options.Response.Challenge.String(), session, time.Minute)
	rsp := &proto.BeginPasskeysLoginResponse{
		Challenge: options.Response.Challenge,
	}
	return rsp, nil
}

// FinishPasskeysLogin implements proto.AuthServer.
func (p *PasskeysService) FinishPasskeysLogin(ctx context.Context, in *proto.FinishPasskeysLoginRequest) (*proto.FinishPasskeysLoginResponse, error) {
	challenge := protocol.URLEncodedBase64(in.Challenge)
	session, found := p.loginSessions.Get(challenge.String())
	if !found {
		return nil, status.Error(codes.InvalidArgument, "会话不存在。")
	}
	defer p.loginSessions.Delete(challenge.String())

	/*
	   {
	   	"id": "lgSUMuk...",
	   	"type": "public-key",
	   	"rawId": "lgSUMuk...",
	   	"response": {
	   		"clientDataJSON": "eyJjaGFsbGVuZ2UiOiJDNn...",
	   		"authenticatorData": "XshXFbsPfvUfduL5wa_7R...",
	   		"signature": "MEQCIG2oLa9WNJJlCUapz8-f22gzfMC...",
	   		"userHandle": "AgAAAA"
	   	}
	   }
	*/
	clientData := protocol.URLEncodedBase64{}
	if err := clientData.UnmarshalJSON([]byte(`"` + in.ClientDataJson + `"`)); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	body := protocol.CredentialAssertionResponse{
		PublicKeyCredential: protocol.PublicKeyCredential{
			Credential: protocol.Credential{
				ID:   protocol.URLEncodedBase64(in.Id).String(),
				Type: `public-key`,
			},
			RawID: in.Id,
		},
		AssertionResponse: protocol.AuthenticatorAssertionResponse{
			AuthenticatorResponse: protocol.AuthenticatorResponse{
				ClientDataJSON: clientData,
			},
			AuthenticatorData: protocol.URLEncodedBase64(in.AuthenticatorData),
			Signature:         protocol.URLEncodedBase64(in.Signature),
			UserHandle:        protocol.URLEncodedBase64(in.UserId),
		},
	}

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader(utils.DropLast1(json.Marshal(body))))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var outUser *user.User
	credential, err := p.wa().FinishDiscoverableLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			id := binary.LittleEndian.Uint32(userHandle)
			u, err := p.authBackend.GetUserByID(int(id))
			if err != nil {
				return nil, err
			}
			outUser = u
			return auth_webauthn.ToWebAuthnUser(outUser), nil
		},
		*session,
		req,
	)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_ = credential

	return &proto.FinishPasskeysLoginResponse{
		Token:   fmt.Sprintf(`%d:%s`, outUser.ID, outUser.Password),
		Cookies: p.cookieGen(outUser, in.UserAgent),
	}, nil
}

var _ proto.PasskeysServer = (*PasskeysService)(nil)
