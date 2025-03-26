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
	//go:embed admin.html chanify.md guest.html
	_root embed.FS

	adminTmpl   = template.Must(template.New(`admin`).ParseFS(_root, `admin.html`)).Lookup(`admin.html`)
	guestTmpl   = template.Must(template.New(`guest`).ParseFS(_root, `guest.html`)).Lookup(`guest.html`)
	chanifyTmpl = text_template.Must(text_template.New(`chanify`).ParseFS(_root, `chanify.md`)).Lookup(`chanify.md`)

	adminPrefix  = `[新的评论]`
	guestPrefix  = `[回复评论]`
	mailFromName = `博客评论`
)

// TODO prettify 内容。
type AdminData struct {
	Title   string `yaml:"title"`
	Link    string `yaml:"link"`
	Date    string `yaml:"date"`
	Author  string `yaml:"author"`
	Content string `yaml:"content"`

	Email    string `yaml:"email"`
	HomePage string `yaml:"home_page"`
}

// TODO prettify 内容。
type GuestData struct {
	Title   string
	Link    string
	Date    string
	Author  string
	Content string
}

type CommentNotifier struct {
	notifier proto.NotifyServer
}

func New(notifier proto.NotifyServer) *CommentNotifier {
	return &CommentNotifier{
		notifier: notifier,
	}
}

func (cn *CommentNotifier) NotifyPostAuthor(data *AdminData, name string, email string) {
	subject := fmt.Sprintf(`%s %s`, adminPrefix, data.Title)

	buf := bytes.NewBuffer(nil)

	buf.Reset()
	chanifyTmpl.Execute(buf, data)
	cn.notifier.SendInstant(
		auth.SystemForLocal(context.Background()),
		&proto.SendInstantRequest{
			Subject: subject,
			Body:    buf.String(),
		},
	)

	buf.Reset()
	adminTmpl.Execute(buf, data)
	cn.notifier.SendEmail(
		auth.SystemForLocal(context.Background()),
		&proto.SendEmailRequest{
			Subject:  subject,
			Body:     buf.String(),
			FromName: mailFromName,
			Users: []*proto.SendEmailRequest_User{
				{
					Name:    name,
					Address: email,
				},
			},
		},
	)
}

func (cn *CommentNotifier) NotifyGuests(data *GuestData, names []string, recipients []string) {
	buf := bytes.NewBuffer(nil)
	if err := guestTmpl.Execute(buf, data); err != nil {
		panic(err)
	}
	subject := fmt.Sprintf(`%s %s`, guestPrefix, data.Title)
	for i := range names {
		cn.notifier.SendEmail(
			auth.SystemForLocal(context.Background()),
			&proto.SendEmailRequest{
				Subject:  subject,
				Body:     buf.String(),
				FromName: mailFromName,
				Users: []*proto.SendEmailRequest_User{
					{
						Name:    names[i],
						Address: recipients[i],
					},
				},
			},
		)
	}
}
