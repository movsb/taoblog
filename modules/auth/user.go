package auth

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/models"
)

// User entity.
type User struct {
	*models.User
}

// 是否为非登录用户。
func (u *User) IsGuest() bool {
	return u == nil || u.ID == 0
}

func (u *User) IsAdmin() bool {
	return u.ID == int64(AdminID)
}

func (u *User) IsSystem() bool {
	return u.ID == int64(systemID)
}

func (u *User) tokenValue() string {
	if u.IsSystem() {
		return fmt.Sprintf(`%d:%s`, u.ID, systemKey)
	}
	return fmt.Sprintf(`%d:%s`, u.ID, u.Password)
}

var _ webauthn.User = (*WebAuthnUser)(nil)

type WebAuthnUser struct {
	*User
	credentials []webauthn.Credential
}

func (u *WebAuthnUser) WebAuthnID() []byte {
	if u.ID <= 0 || u.ID > math.MaxInt32 {
		panic(`user id is invalid`)
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(u.ID))
	return buf
}
func (u *WebAuthnUser) WebAuthnName() string {
	return fmt.Sprintf(`%s (id:%d)`, u.Nickname, u.ID)
}
func (u *WebAuthnUser) WebAuthnDisplayName() string {
	return u.Nickname
}
func (u *WebAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}
func (u *WebAuthnUser) WebAuthnIcon() string {
	return ""
}

func randomKey() string {
	b := [32]byte{}
	utils.Must1(rand.Read(b[:]))
	return fmt.Sprintf(`%x`, b)
}

var (
	// TODO 移除，用 nil 代表未登录用户。
	guest = &User{
		User: &models.User{
			ID:       0,
			Nickname: `未登录用户`,
		},
	}
	// TODO 怎么确保程序重启后一定不一样？
	systemKey = randomKey()
	systemID  = 1
	system    = &User{
		User: &models.User{
			ID: int64(systemID),
		},
	}
	AdminID = 2
)

// 仅能同进程内使用。
func SystemToken() string {
	return system.tokenValue()
}
