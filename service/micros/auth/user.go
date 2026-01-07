package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/db"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/cookies"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm"
	"github.com/phuslu/lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserManager struct {
	ctx context.Context
	db  *taorm.DB

	userCache *lru.LRUCache[int, *models.User]

	proto.UnimplementedUsersServer
}

func NewUsersService(ctx context.Context, db *taorm.DB, sr grpc.ServiceRegistrar) *UserManager {
	um := &UserManager{
		ctx: ctx,
		db:  db,

		userCache: lru.NewLRUCache[int, *models.User](8),
	}
	proto.RegisterUsersServer(sr, um)
	return um
}

func (m *UserManager) dropCache(id int) {
	m.userCache.Delete(id)
}

func (m *UserManager) CreateUser(ctx context.Context, in *proto.User) (*proto.User, error) {
	user.MustBeAdmin(ctx)

	var password [16]byte
	utils.Must1(rand.Read(password[:]))
	passwordString := fmt.Sprintf(`%x`, password)

	now := time.Now()

	user := models.User{
		CreatedAt: now.Unix(),
		UpdatedAt: now.Unix(),
		Nickname:  strings.TrimSpace(in.Nickname),
		Password:  passwordString,
		Hidden:    in.Hidden,
	}

	if user.Nickname == `` {
		return nil, status.Error(codes.InvalidArgument, `昵称不能为空。`)
	}

	if err := m.db.Model(&user).Create(); err != nil {
		return nil, err
	}

	// userCache 会缓存不存在的用户，这里也要清理。
	m.dropCache(int(user.ID))

	return user.ToProto(), nil
}

func (m *UserManager) UpdateUser(ctx context.Context, in *proto.UpdateUserRequest) (_ *proto.UpdateUserResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	ac := user.Context(ctx)
	if !(ac.User.IsAdmin() || ac.User.IsSystem() || (in.User.Id > 0 && ac.User.ID == in.User.Id)) {
		return nil, status.Error(codes.PermissionDenied, `没有权限完成此操作。`)
	}

	mm := taorm.M{}

	if in.UpdateAvatar {
		d := utils.Must1(utils.ParseDataURL(in.User.Avatar))
		if !strings.HasPrefix(d.Type, `image/`) {
			panic(`不是图片文件。`)
		}
		if len(d.Data) > 10<<20 {
			return nil, status.Error(codes.InvalidArgument, `图片太大。`)
		}
		if len(d.Data) > 220<<10 {
			d, err := utils.ResizeImage(d.Type, bytes.NewReader(d.Data), 220, 220)
			if err != nil {
				return nil, err
			}
			mm[`avatar`] = d.String()
		} else {
			mm[`avatar`] = in.User.Avatar
		}
	}
	if in.UpdateEmail {
		if in.User.Email != `` && !utils.IsEmail(in.User.Email) {
			panic(status.Error(codes.InvalidArgument, `邮箱错误。`))
		}
		mm[`email`] = in.User.Email
	}
	if in.UpdateBarkToken {
		if in.User.BarkToken != `` && len(in.User.BarkToken) > 1024 {
			panic(status.Error(codes.InvalidArgument, `Bark Token 错误。`))
		}
		mm[`bark_token`] = in.User.BarkToken
	}
	if in.UpdateNickname {
		mm[`nickname`] = in.User.Nickname
	}

	if len(mm) > 0 {
		r := m.db.Model(models.User{ID: in.User.Id}).MustUpdateMap(mm)
		n := utils.Must1(r.RowsAffected())
		if n != 1 {
			panic(`更新失败。`)
		}
		m.dropCache(int(in.User.Id))
	}

	return &proto.UpdateUserResponse{}, nil
}

func (m *UserManager) ListUsers(ctx context.Context, in *proto.ListUsersRequest) (_ *proto.ListUsersResponse, outErr error) {
	defer utils.CatchAsError(&outErr)

	user.MustBeAdmin(ctx)

	var users []*models.User

	stmt := m.db.Select(`*`)

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

// NOTE: 错误的时候也会缓存，nil 值，以避免不必要的查询。
func (m *UserManager) GetUserByID(ctx context.Context, id int) (*user.User, error) {
	// 系统是虚的，不在数据库内。
	if id == user.SystemID {
		return user.System, nil
	}

	mu, err, _ := m.userCache.GetOrLoad(ctx, int(id), func(ctx context.Context, i int) (*models.User, error) {
		var user models.User
		err := db.FromContextDefault(ctx, m.db).Where(`id=?`, id).Find(&user)
		if err != nil {
			if !taorm.IsNotFoundError(err) {
				return nil, fmt.Errorf(`GetUserByID: %w`, err)
			}
			// NOTE: 错误的时候也会存，nil 值，以避免不必要的查询。
			// 但是最终会返回错误；CreateUser 的时候清除。
			return nil, nil
		}
		return &user, nil
	})

	if mu == nil {
		err = sql.ErrNoRows
	}

	return &user.User{User: mu}, err
}

// https://x.com/sebastienlorber/status/1932367017065025675
func constantEqual(x, y string) bool {
	return subtle.ConstantTimeCompare([]byte(x), []byte(y)) == 1
}

// 通过比对密码返回用户。
func (m *UserManager) GetUserByPassword(ctx context.Context, id int, password string) (*user.User, error) {
	u, err := m.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if constantEqual(u.Password, password) {
		return u, nil
	}
	return nil, sql.ErrNoRows
}

// 通过比对凭证返回用户。
func (m *UserManager) GetUserByToken(ctx context.Context, id int, token string) (*user.User, error) {
	u, err := m.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if constantEqual(token, cookies.TokenValue(id, u.Password)) {
		return u, nil
	}
	return nil, sql.ErrNoRows
}

// proto 里面的用户不会返回此 secret，所以新开函数。
func (m *UserManager) SetUserOTPSecret(user *user.User, secret string) {
	m.db.Model(user.User).Where(`id=?`, user.ID).MustUpdateMap(taorm.M{
		`otp_secret`: secret,
	})
	m.dropCache(int(user.ID))
}

func (m *UserManager) AddWebAuthnCredential(user *user.User, cred *webauthn.Credential) {
	existed := slices.IndexFunc(user.Credentials, func(c webauthn.Credential) bool {
		return bytes.Equal(c.PublicKey, cred.PublicKey) ||
			// 不允许为同一认证器添加多个凭证。
			// TODO 认证器为了隐私会使 AAGUID 全部为零，这里的判断无效。
			bytes.Equal(c.Authenticator.AAGUID, cred.Authenticator.AAGUID)
	})
	if existed >= 0 {
		user.Credentials[existed] = *cred
	} else {
		user.Credentials = append(user.Credentials, *cred)
	}
	m.db.Model(user.User).Where(`id=?`, user.ID).MustUpdateMap(taorm.M{
		`credentials`: user.Credentials,
	})
	m.dropCache(int(user.ID))
}
