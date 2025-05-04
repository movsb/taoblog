package comment_notify

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	text_template "text/template"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols/go/proto"
)

var (
	//go:embed author.html chanify.md guest.html
	_root embed.FS

	authorTmpl  = template.Must(template.New(`author`).ParseFS(_root, `author.html`)).Lookup(`author.html`)
	guestTmpl   = template.Must(template.New(`guest`).ParseFS(_root, `guest.html`)).Lookup(`guest.html`)
	chanifyTmpl = text_template.Must(text_template.New(`chanify`).ParseFS(_root, `chanify.md`)).Lookup(`chanify.md`)

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

func (cn *CommentNotifier) NotifyAdmin(d *AdminData) {
	buf := bytes.NewBuffer(nil)
	chanifyTmpl.Execute(buf, d)
	cn.notifier.SendInstant(
		auth.SystemForLocal(context.Background()),
		&proto.SendInstantRequest{
			Subject: fmt.Sprintf(`%s%s`, authorPrefix, d.Title),
			Body:    buf.String(),
		},
	)
}

type Recipient struct {
	Name    string
	Address string
}

func (cn *CommentNotifier) NotifyPostAuthor(data *Data, name string, email string) {
	buf := bytes.NewBuffer(nil)
	authorTmpl.Execute(buf, data)
	subject := fmt.Sprintf(`%s %s`, authorPrefix, data.Title)
	cn.sendEmails(subject, buf.String(), []Recipient{{Name: name, Address: email}})
}

func (cn *CommentNotifier) NotifyGuests(data *Data, recipients []Recipient) {
	buf := bytes.NewBuffer(nil)
	guestTmpl.Execute(buf, data)
	subject := fmt.Sprintf(`%s %s`, guestPrefix, data.Title)
	cn.sendEmails(subject, buf.String(), recipients)
}

// 重要！ 每封邮件必须独立发送 （To:），否则会相互看到别人地址、所有人的地址。
func (cn *CommentNotifier) sendEmails(subject, body string, recipients []Recipient) {
	for _, r := range recipients {
		cn.sendEmail(subject, body, r)
	}
}

func (cn *CommentNotifier) sendEmail(subject, body string, recipient Recipient) {
	cn.notifier.SendEmail(
		auth.SystemForLocal(context.Background()),
		&proto.SendEmailRequest{
			Subject:  subject,
			Body:     body,
			FromName: mailFromName,
			Users: []*proto.SendEmailRequest_User{
				{
					Name:    recipient.Name,
					Address: recipient.Address,
				},
			},
		},
	)
}
