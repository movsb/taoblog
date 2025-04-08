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
	dropCache func(id int)

	loginSessions *lru.TTLCache[string, *webauthn.SessionData]

	proto.UnimplementedAuthServer
}

func NewPasskeys(
	db *taorm.DB,
	wa *webauthn.WebAuthn,
	cookieGen func(user *User, agent string) []*proto.FinishPasskeysLoginResponse_Cookie,
	dropCache func(id int),
) *Passkeys {
	return &Passkeys{
		db:            db,
		wa:            wa,
		cookieGen:     cookieGen,
		dropCache:     dropCache,
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
	MustBeAdmin(ctx)

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

	// userCache 会缓存不存在的用户，这里也要清理。
	p.dropCache(int(user.ID))

	return user.ToProto(), nil
}

func (p *Passkeys) UpdateUser(ctx context.Context, in *proto.UpdateUserRequest) (_ *proto.UpdateUserResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := Context(ctx)
	if !(ac.User.IsAdmin() || ac.User.IsSystem() || (in.User.Id > 0 && ac.User.ID == in.User.Id)) {
		panic(noPerm)
	}

	m := taorm.M{}

	if in.UpdateAvatar {
		d := utils.Must1(utils.ParseDataURL(in.User.Avatar))
		if len(d.Data) > 100<<10 {
			panic(`头像太大。`)
		}
		if !strings.HasPrefix(d.Type, `image/`) {
			panic(`不是图片文件。`)
		}
		m[`avatar`] = in.User.Avatar
	}

	if len(m) > 0 {
		r := p.db.Model(models.User{ID: in.User.Id}).MustUpdateMap(m)
		n := utils.Must1(r.RowsAffected())
		if n != 1 {
			panic(`更新失败。`)
		}
		p.dropCache(int(in.User.Id))
	}

	return &proto.UpdateUserResponse{}, nil
}

func (p *Passkeys) ListUsers(ctx context.Context, in *proto.ListUsersRequest) (_ *proto.ListUsersResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	MustBeAdmin(ctx)

	var users []*models.User

	stmt := p.db.Select(`*`)

	if !in.WithUnnamed {
		stmt = stmt.Where(`nickname!=''`)
	}

	if !in.WithHidden {
		stmt = stmt.Where(`hidden=0`)
	}

	stmt.MustFind(&users)

	return &proto.ListUsersResponse{
		Users: utils.Map(users, func(u *models.User) *proto.User { return u.ToProto() }),
	}, nil
}
