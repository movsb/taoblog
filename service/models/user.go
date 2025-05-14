package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
)

// NOTE: 很多字段并没有输出到 proto 中。
type User struct {
	ID        int64
	CreatedAt int64
	UpdatedAt int64
	Nickname  string
	Password  string

	/// 社交账号绑定信息

	// 邮箱地址。
	// 并非必须。
	Email string

	// Passkeys 凭证。
	Credentials Credentials

	GoogleUserID string

	// 小写 GitHub 是为了使默认数据库字段名为：github_user_id
	// 否则默认的蛇形规则可能是 git_hub_user_id
	GithubUserID string

	Hidden bool

	Avatar Avatar

	// 一次性认证 secret。
	// 通用算法/长度，所以只保存了 Secret。
	OtpSecret string

	ChanifyToken string
}

type Avatar utils.DataURL

func (a Avatar) Value() (driver.Value, error) {
	return utils.DataURL(a).String(), nil
}
func (a *Avatar) Scan(v any) error {
	var d []byte
	switch val := v.(type) {
	case string:
		d = []byte(val)
	case []byte:
		d = val
	}
	if len(d) == 0 {
		return nil
	}
	u, err := utils.ParseDataURL(string(d))
	*a = Avatar(*u)
	return err
}

type Credentials []webauthn.Credential

func (c Credentials) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *Credentials) Scan(v any) error {
	switch val := v.(type) {
	case string:
		// TODO 支持 NULL，以处理空字符串。
		if val == `` {
			return nil
		}
		return json.Unmarshal([]byte(val), c)
	case []byte:
		return json.Unmarshal(val, c)
	}
	return errors.New(`unsupported type for credentials`)
}

func (User) TableName() string {
	return `users`
}

func (u *User) ToProto() *proto.User {
	return &proto.User{
		Id:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Nickname:  u.Nickname,
		Password:  u.Password,

		Email:        u.Email,
		ChanifyToken: u.ChanifyToken,
		// hidden
		// otp_secret
	}
}

type Perm string

const (
	PermRead = `read`
)

type AccessControlEntry struct {
	ID         int64
	CreatedAt  int64
	UserID     int64
	PostID     int64
	Permission string
}

func (AccessControlEntry) TableName() string {
	return `acl`
}
