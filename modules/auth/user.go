package auth

import (
	"encoding/binary"
	"math"

	"github.com/go-webauthn/webauthn/webauthn"
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

var (
	guest = &User{
		ID:                  0,
		Email:               "",
		DisplayName:         "未注册用户",
		webAuthnCredentials: nil,
	}
	admin = &User{
		ID: 1,
	}
)
