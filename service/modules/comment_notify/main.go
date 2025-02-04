package comment_notify

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	text_template "text/template"

	"github.com/movsb/taoblog/modules/mailer"
	"github.com/movsb/taoblog/modules/notify"
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
	mailer   *mailer.MailerLogger
	notifier notify.Notifier
}

func New(notifier notify.Notifier, mailer *mailer.MailerLogger) *CommentNotifier {
	return &CommentNotifier{
		mailer:   mailer,
		notifier: notifier,
	}
}

func (cn *CommentNotifier) NotifyPostAuthor(data *AdminData, name string, email string) {
	subject := fmt.Sprintf(`%s %s`, adminPrefix, data.Title)

	buf := bytes.NewBuffer(nil)

	buf.Reset()
	chanifyTmpl.Execute(buf, data)
	cn.notifier.Notify(subject, buf.String())

	buf.Reset()
	adminTmpl.Execute(buf, data)
	cn.mailer.Queue(
		subject, buf.String(), mailFromName,
		[]mailer.User{
			{
				Name:    name,
				Address: email,
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
	body := buf.String()
	for i := 0; i < len(names); i++ {
		cn.mailer.Queue(
			subject, body, mailFromName,
			[]mailer.User{
				{
					Name:    names[i],
					Address: recipients[i],
				},
			},
		)
	}
}
