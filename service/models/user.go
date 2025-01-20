package models

import "github.com/movsb/taoblog/protocols/go/proto"

type User struct {
	ID        int64
	CreatedAt int64
	UpdatedAt int64
	Password  string
}

func (User) TableName() string {
	return `users`
}

func (u *User) ToProto() *proto.User {
	return &proto.User{
		Id:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Password:  u.Password,
	}
}

type SocialBinding struct {
	ID          int64
	CreatedAt   int64
	UserID      int64
	Type        string
	SocialID    string
	SocialLogin string
}

func (SocialBinding) TableName() string {
	return `social_bindings`
}

type AccessControlEntry struct {
	ID         int64
	CreatedAt  int64
	UserID     int64
	PostID     int64
	Permission string
}
