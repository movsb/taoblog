package comment_notify

import (
	"bytes"
	"html/template"
	"log"

	"github.com/movsb/taoblog/modules/mailer"
)

var adminTmpl *template.Template
var guestTmpl *template.Template

func init() {
	adminTmpl = template.Must(template.New("admin").Parse(adminTemplate))
	guestTmpl = template.Must(template.New("guest").Parse(guestTemplate))
}

type AdminData struct {
	Title   string
	Link    string
	Date    string
	Author  string
	Content string

	Email    string
	HomePage string
}

type GuestData struct {
	Title   string
	Link    string
	Date    string
	Author  string
	Content string
}

type CommentNotifier struct {
	AdminName  string
	AdminEmail string
	MailServer string
	Username   string
	Password   string
}

func (cn *CommentNotifier) NotifyAdmin(data *AdminData) {
	buf := bytes.NewBuffer(nil)
	if err := adminTmpl.Execute(buf, data); err != nil {
		panic(err)
	}
	subject := "[新博文评论] " + data.Title
	body := buf.String()
	cn.sendMailAsync(cn.AdminName, cn.AdminEmail, subject, body)
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
	log.Printf("SendMail: %s[%s] - %s", recipientName, recipientAddress, subject)
	go func() {
		for i := 0; i < 3; i++ {
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
			break
		}
	}()
}
