package comment_notify

import (
	"bytes"
	"html/template"
	"log"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/mailer"
)

var adminTmpl *template.Template
var guestTmpl *template.Template

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
	MailServer string
	Username   string
	Password   string
	Config     *config.CommentConfig
}

func (cn *CommentNotifier) Init() {
	adminTmpl = template.Must(template.New("admin").Parse(cn.Config.Templates.Admin))
	guestTmpl = template.Must(template.New("guest").Parse(cn.Config.Templates.Guest))
}

func (cn *CommentNotifier) NotifyAdmin(data *AdminData) {
	buf := bytes.NewBuffer(nil)
	if err := adminTmpl.Execute(buf, data); err != nil {
		panic(err)
	}
	subject := "[新博文评论] " + data.Title
	body := buf.String()
	cn.sendMailAsync(cn.Config.Author, cn.Config.Emails[0], subject, body)
}

func (cn *CommentNotifier) NotifyGuests(data *GuestData, names []string, recipients []string) {
	buf := bytes.NewBuffer(nil)
	if err := guestTmpl.Execute(buf, data); err != nil {
		panic(err)
	}
	subject := "[回复评论] " + data.Title
	body := buf.String()
	for i := 0; i < len(names); i++ {
		cn.sendMailAsync(names[i], recipients[i], subject, body)
	}
}

func (cn *CommentNotifier) sendMailAsync(
	recipientName string, recipientAddress string,
	subject string, body string,
) {
	log.Printf("SendMail: %s[%s] - %s\n\n%s\n\n", recipientName, recipientAddress, subject, body)

	if !cn.Config.Notify {
		log.Println(`邮件被禁用，将不发送。`)
		return
	}

	go func() {
		mc, err := mailer.DialTLS(cn.MailServer)
		if err != nil {
			log.Println(err)
			return
		}
		defer mc.Quit()
		if err = mc.Auth(cn.Username, cn.Password); err != nil {
			log.Println(err)
			return
		}
		if err = mc.SetFrom("博客评论", cn.Username); err != nil {
			log.Println("SetFrom:", err)
			return
		}
		if err = mc.AddTo(recipientName, recipientAddress); err != nil {
			log.Println("AddTo:", recipientAddress, err)
			return
		}
		if err = mc.Send(subject, body); err != nil {
			log.Println(err)
			return
		}
	}()
}
