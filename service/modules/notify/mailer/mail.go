package mailer

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/logs"
	"github.com/movsb/taoblog/modules/utils"
)

type MailSender interface {
	Send(subject, body string, fromName string, tos []User) error
}

type Mailer struct {
	server             string
	username, password string
	auth               smtp.Auth
}

func NewMailer(server string, username, password string) Mailer {
	host, _, _ := net.SplitHostPort(server)
	return Mailer{
		server:   server,
		username: username,
		password: password,
		auth:     smtp.PlainAuth(``, username, password, host),
	}
}

type User struct {
	Name    string
	Address string
}

func (u User) String() string {
	return fmt.Sprintf(`%s <%s>`, encode(u.Name), u.Address)
}

func encode(s string) string {
	return mime.BEncoding.Encode(`utf-8`, string(s))
}

func (c *Mailer) Send(subject, body string, fromName string, tos []User) error {
	mailBody := writeBody(subject, body, User{Name: fromName, Address: c.username}, tos)
	to := utils.Map(tos, func(u User) string { return u.Address })
	err := smtp.SendMail(c.server, c.auth, c.username, to, mailBody)
	// fix for qq mail
	if err != nil {
		if strings.HasSuffix(err.Error(), "\x00\x00\x00\x1a\x00\x00\x00") {
			err = nil
		}
	}
	return err
}

func writeBody(subject, body string, from User, tos []User) []byte {
	w := bytes.NewBuffer(nil)
	f := func(ss ...string) {
		for _, s := range ss {
			fmt.Fprint(w, s)
		}
		fmt.Fprint(w, "\r\n")
	}
	f(`Subject: `, subject)
	f(`From: `, from.String())
	for _, to := range tos {
		f(`To: `, to.String())
	}
	f(`Content-Type: text/html; charset=utf-8`)
	f()
	f(body)
	return w.Bytes()
}

type MailerLogger struct {
	store logs.Logger

	mailer       Mailer
	pullInterval time.Duration
}

func NewMailerLogger(store logs.Logger, mailer Mailer) *MailerLogger {
	n := &MailerLogger{
		store:  store,
		mailer: mailer,
	}
	n.SetPullInterval(time.Minute)
	go n.process(context.Background())
	return n
}

const (
	ty  = `mail`
	sty = `message`
)

type _Message struct {
	// From Address 始终使用 username
	FromName string
	Tos      []User
	Subject  string
	Body     string
}

func (n *MailerLogger) SetPullInterval(d time.Duration) {
	if d <= time.Second*10 {
		d = time.Second * 10
	}
	n.pullInterval = d
}

func (n *MailerLogger) Send(subject, body string, fromName string, tos []User) error {
	n.store.CreateLog(context.Background(), ty, sty, 1, _Message{
		FromName: fromName,
		Tos:      tos,
		Subject:  subject,
		Body:     body,
	})
	return nil
}

func (n *MailerLogger) process(ctx context.Context) {
	for {
		var msg _Message
		l := n.store.FindLog(ctx, ty, sty, &msg)
		if l != nil {
			log.Println(`找到日志：`, l.ID)
			if err := n.mailer.Send(msg.Subject, msg.Body, msg.FromName, msg.Tos); err != nil {
				log.Println(`MailError:`, err)

				// 有些错误不可重试
				// ChatGPT：smtp 错误码？
				//
				// 白痴QQ邮箱连续发送几封邮件就报未登录错误，宛若智障。
				// MailError: 535 Login Fail. Please enter your authorization code to login.
				// More information in https://service.mail.qq.com/detail/0/53
				// var te *textproto.Error
				// if errors.As(err, &te) {
				// 	if te.Code >= 500 {
				// 		log.Println(`不可重试错误，将删除邮件。`)
				// 		n.store.DeleteLog(ctx, l.ID)
				// 		continue
				// 	}
				// }

				time.Sleep(n.pullInterval)
				continue
			} else {
				n.store.DeleteLog(ctx, l.ID)
			}
		} else {
			time.Sleep(n.pullInterval)
		}
	}
}
