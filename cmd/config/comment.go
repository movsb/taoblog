package config

import (
	"fmt"
	"net"

	"github.com/movsb/taoblog/modules/utils"
)

// CommentConfig ...
type CommentConfig struct {
	Author            string                 `yaml:"author"`
	Emails            []string               `yaml:"emails"`
	Notify            bool                   `yaml:"notify"`
	NotAllowedEmails  []string               `yaml:"not_allowed_emails"`
	NotAllowedAuthors []string               `yaml:"not_allowed_authors"`
	Templates         CommentTemplatesConfig `yaml:"templates"`
}

// DefaultCommentConfig ...
func DefaultCommentConfig() CommentConfig {
	return CommentConfig{
		Templates: DefaultCommentTemplatesConfig(),
	}
}

// CommentTemplatesConfig ...
type CommentTemplatesConfig struct {
	Admin string `yaml:"admin"`
	Guest string `yaml:"guest"`
}

// DefaultCommentTemplatesConfig ...
func DefaultCommentTemplatesConfig() CommentTemplatesConfig {
	const adminTemplate = `
<b>您的博文“{{.Title}}”有新的评论啦！</b><br/><br/>

<pre>{{ .Content }}</pre>
<br/>

<b>链接：</b><a href="{{.Link}}">{{.Link}}</a><br/>
<b>作者：</b>{{.Author}}<br/>
<b>邮箱：</b>{{.Email}}<br/>
<b>网址：</b>{{.HomePage}}<br/>
<b>时间：</b>{{.Date}}<br/>
`

	const guestTemplate = `
<b>您在博文“{{.Title}}”的评论有新的回复啦！</b><br/><br/>

<pre>{{ .Content }}</pre>
<br/>

<b>链接：</b><a href="{{.Link}}">{{.Link}}</a><br/>
<b>作者：</b>{{.Author}}<br/>
<b>时间：</b>{{.Date}}<br/>

<br/>该邮件为系统自动发出，请勿直接回复该邮件。<br/>
`
	return CommentTemplatesConfig{
		Admin: adminTemplate,
		Guest: guestTemplate,
	}
}

type NotificationConfig struct {
	Chanify NotificationChanifyConfig `json:"chanify" yaml:"chanify"`
	Mailer  NotificationMailerConfig  `json:"mailer" yaml:"mailer"`
}

type NotificationChanifyConfig struct {
	Token string `json:"token" yaml:"token"`
}

func (NotificationChanifyConfig) CanSave() {}

func (c *NotificationChanifyConfig) BeforeSet(paths Segments, obj any) error {
	switch paths.At(0).Key {
	case `token`:
		return nil
	}
	return fmt.Errorf(`不存在的设置字段：%s`, paths)
}

type NotificationMailerConfig struct {
	// 服务器地址。形如：smtp.example.com:465
	Server string `json:"server" yaml:"server"`
	// 帐户名。通常是邮件地址。
	Account string `json:"account" yaml:"account"`
	// 密码。
	Password string `json:"password" yaml:"password"`
}

func (NotificationMailerConfig) CanSave() {}

func (c *NotificationMailerConfig) BeforeSet(paths Segments, obj any) error {
	var new NotificationMailerConfig
	if len(paths) == 0 {
		new = obj.(NotificationMailerConfig)
	} else {
		switch paths.At(0).Key {
		case `server`:
			new.Server = obj.(string)
		case `account`:
			new.Account = obj.(string)
		case `password`:
			new.Password = obj.(string)
		}
	}
	if new.Server != "" {
		host, port, err := net.SplitHostPort(new.Server)
		if err != nil {
			return err
		}
		if host == "" || port == "" {
			return fmt.Errorf(`主机或者端口不正确：%q`, new.Server)
		}
	}
	if new.Account != "" {
		if !utils.IsEmail(new.Account) {
			return fmt.Errorf(`看起来不是正确的邮件地址：%q`, new.Account)
		}
	}
	return nil
}
