package user

import (
	"crypto/rand"
	"fmt"

	"github.com/movsb/taoblog/modules/auth/cookies"
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
	return u.ID == int64(SystemID)
}

func (u *User) tokenValue() string {
	return cookies.TokenValue(int(u.ID), u.Password)
}

func (u *User) AuthorizationValue() string {
	return cookies.TokenName + ` ` + u.tokenValue()
}

func randomKey() string {
	b := [32]byte{}
	utils.Must1(rand.Read(b[:]))
	return fmt.Sprintf(`%x`, b)
}

// 必须大于零
const (
	SystemID = 1
	AdminID  = 2
)

var (
	// TODO 移除，用 nil 代表未登录用户。
	Guest = &User{
		User: &models.User{
			ID:       0,
			Nickname: `未登录用户`,
		},
	}
	System = &User{
		User: &models.User{
			ID:       int64(SystemID),
			Password: randomKey(),
		},
	}
)

// 仅能同进程内使用。
func SystemTokenValue() string {
	return System.tokenValue()
}
