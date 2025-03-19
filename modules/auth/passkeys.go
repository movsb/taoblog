package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm"
	"github.com/phuslu/lru"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TODO 更改成 Auth 服务。
type Passkeys struct {
	db        *taorm.DB
	wa        *webauthn.WebAuthn
	cookieGen func(user *User, agent string) []*proto.FinishPasskeysLoginResponse_Cookie

	loginSessions *lru.TTLCache[string, *webauthn.SessionData]

	proto.UnimplementedAuthServer
}

func NewPasskeys(
	db *taorm.DB,
	wa *webauthn.WebAuthn,
	cookieGen func(user *User, agent string) []*proto.FinishPasskeysLoginResponse_Cookie,
) *Passkeys {
	return &Passkeys{
		db:            db,
		wa:            wa,
		cookieGen:     cookieGen,
		loginSessions: lru.NewTTLCache[string, *webauthn.SessionData](8),
	}
}

// BeginPasskeysLogin implements proto.AuthServer.
func (p *Passkeys) BeginPasskeysLogin(context.Context, *proto.BeginPasskeysLoginRequest) (*proto.BeginPasskeysLoginResponse, error) {
	options, session, err := p.wa.BeginDiscoverableLogin()
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
func (p *Passkeys) FinishPasskeysLogin(ctx context.Context, in *proto.FinishPasskeysLoginRequest) (*proto.FinishPasskeysLoginResponse, error) {
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

	var outUser *User
	credential, err := p.wa.FinishDiscoverableLogin(
		func(rawID, userHandle []byte) (webauthn.User, error) {
			id := binary.LittleEndian.Uint32(userHandle)
			var user models.User
			if err := p.db.Where(`id=?`, id).Find(&user); err != nil {
				return nil, err
			}
			outUser = &User{&user}
			return ToWebAuthnUser(outUser), nil
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

var _ proto.AuthServer = (*Passkeys)(nil)

func ToWebAuthnUser(u *User) webauthn.User {
	return &WebAuthnUser{
		User:        u,
		credentials: u.Credentials,
	}
}

func (p *Passkeys) CreateUser(ctx context.Context, in *proto.User) (*proto.User, error) {
	var password [16]byte
	utils.Must1(rand.Read(password[:]))
	passwordString := fmt.Sprintf(`%x`, password)

	now := time.Now()

	user := models.User{
		CreatedAt: now.Unix(),
		UpdatedAt: now.Unix(),
		Nickname:  strings.TrimSpace(in.Nickname),
		Password:  passwordString,
	}

	if user.Nickname == `` {
		return nil, status.Error(codes.InvalidArgument, `昵称不能为空。`)
	}

	if err := p.db.Model(&user).Create(); err != nil {
		return nil, err
	}

	return user.ToProto(), nil
}
