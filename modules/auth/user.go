package auth

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/movsb/taoblog/modules/utils"
)

// User entity.
type User struct {
	ID          int64  // 不可变 ID
	Email       string // 可变 ID
	DisplayName string // 昵称

	webAuthnCredentials []webauthn.Credential
}

// IsGuest ...
func (u *User) IsGuest() bool {
	return u == nil || u.ID == 0
}

// IsAdmin ...
func (u *User) IsAdmin() bool {
	return u.ID == admin.ID
}

func (u *User) IsSystem() bool {
	return u.ID == system.ID
}

var _ webauthn.User = (*User)(nil)

func (u *User) WebAuthnID() []byte {
	if u.ID <= 0 || u.ID > math.MaxInt32 {
		panic(`user id is invalid`)
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(u.ID))
	return buf
}
func (u *User) WebAuthnName() string {
	return u.Email
}
func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.webAuthnCredentials
}
func (u *User) WebAuthnIcon() string {
	return ""
}

func randomKey() string {
	b := [32]byte{}
	utils.Must1(rand.Read(b[:]))
	return fmt.Sprintf(`%x`, b)
}

var (
	guest = &User{
		ID:                  0,
		Email:               "",
		DisplayName:         "未注册用户",
		webAuthnCredentials: nil,
	}
	// TODO 怎么确保程序重启后一定不一样？
	systemKey = randomKey()
	system    = &User{
		ID: 1,
	}
	AdminID = 2
	admin   = &User{
		ID: int64(AdminID),
	}
	Notify = &User{
		ID: 3,
	}
)
