package comment_notify

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	text_template "text/template"

	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
)

var (
	//go:embed author.html bark.md guest.html
	_root embed.FS

	authorMailTmpl = template.Must(template.New(`author`).ParseFS(_root, `author.html`)).Lookup(`author.html`)
	guestMailTmpl  = template.Must(template.New(`guest`).ParseFS(_root, `guest.html`)).Lookup(`guest.html`)
	barkTmpl       = text_template.Must(text_template.New(`bark`).ParseFS(_root, `bark.md`)).Lookup(`bark.md`)
)

const (
	authorPrefix = `[新的评论]`
	guestPrefix  = `[回复评论]`
	mailFromName = `文章评论`
)

type Data struct {
	Title   string
	Link    string
	Date    string
	Author  string
	Content string
}

type AdminData struct {
	Data

	Email    string
	HomePage string
}

type CommentNotifier struct {
	notifier proto.NotifyServer
}

func New(notifier proto.NotifyServer) *CommentNotifier {
	return &CommentNotifier{
		notifier: notifier,
	}
}

// 这是给站长的通知。
//
// 会收到所有评论的通知。
// 只发送即时通知，不发送邮件。
func (cn *CommentNotifier) NotifyAdmin(d *AdminData) {
	buf := bytes.NewBuffer(nil)
	barkTmpl.Execute(buf, d)
	cn.sendNotify(
		fmt.Sprintf(`%s%s`, authorPrefix, d.Title),
		buf.String(),
		``,
	)
}

type Recipient struct {
	Email struct {
		Name    string
		Address string
	}
	BarkToken string
}

// 通知文章作者。
//
//   - 如果有配置即时通知，只发送即时通知。
//   - 如果有配置邮件通知，只发送邮件通知。
func (cn *CommentNotifier) NotifyPostAuthor(data *Data, recipient Recipient) {
	subject := fmt.Sprintf(`%s %s`, authorPrefix, data.Title)
	body := bytes.NewBuffer(nil)
	if recipient.BarkToken != `` {
		barkTmpl.Execute(body, data)
		cn.sendNotify(subject, body.String(), recipient.BarkToken)
	} else if recipient.Email.Address != `` {
		authorMailTmpl.Execute(body, data)
		cn.sendEmail(subject, body.String(), recipient.Email.Name, recipient.Email.Address)
	} else {
		log.Println(`文章作者没有通知功能设置，将不发送任何通知。`)
	}
}

// 所有除站长和文章本人以外的评论者。
func (cn *CommentNotifier) NotifyGuests(data *Data, recipients []Recipient) {
	subject := fmt.Sprintf(`%s %s`, guestPrefix, data.Title)

	body1 := bytes.NewBuffer(nil)
	body2 := bytes.NewBuffer(nil)
	barkTmpl.Execute(body1, data)
	guestMailTmpl.Execute(body2, data)

	// 重要！ 每封邮件必须独立发送 （To:），否则会相互看到别人地址、所有人的地址。
	for _, r := range recipients {
		if r.BarkToken != `` {
			cn.sendNotify(subject, body1.String(), r.BarkToken)
		} else if r.Email.Address != `` {
			cn.sendEmail(subject, body2.String(), r.Email.Name, r.Email.Address)
		} else {
			log.Println(`评论者没有通知功能设置，将不发送任何通知。`)
		}
	}
}

func (cn *CommentNotifier) sendNotify(subject, body string, token string) {
	cn.notifier.SendInstant(
		user.SystemForLocal(context.Background()),
		&proto.SendInstantRequest{
			Title:     subject,
			Body:      body,
			BarkToken: token,
		},
	)
}

func (cn *CommentNotifier) sendEmail(subject, body string, name, address string) {
	cn.notifier.SendEmail(
		user.SystemForLocal(context.Background()),
		&proto.SendEmailRequest{
			Subject:  subject,
			Body:     body,
			FromName: mailFromName,
			Users: []*proto.SendEmailRequest_User{
				{
					Name:    name,
					Address: address,
				},
			},
		},
	)
}
