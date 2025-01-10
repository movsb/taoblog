package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/smtp"
	"net/textproto"
	"time"

	"github.com/movsb/taoblog/modules/logs"
)

// Mailer is a SMTP mail sender.
type Mailer struct {
	conn     net.Conn
	c        *smtp.Client
	host     string
	fromMail string
	fromName string
	toEmails []string
	toNames  []string
}

// DialTLS dials mail server.
func DialTLS(host string) (*Mailer, error) {
	conn, err := tls.Dial("tcp", host, nil)
	if err != nil {
		return nil, err
	}
	h, _, _ := net.SplitHostPort(host)
	return New(conn, h)
}

// conn 必须是 tls.Conn
func New(conn net.Conn, host string) (*Mailer, error) {
	if _, ok := conn.(*tls.Conn); !ok {
		panic(`not tls`)
	}
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		conn.Close()
		return nil, err
	}

	return &Mailer{
		host: host,
		conn: conn,
		c:    c,
	}, nil
}

// Auth authenticates
func (m *Mailer) Auth(username string, password string) error {
	auth := smtp.PlainAuth("", username, password, m.host)
	return m.c.Auth(auth)
}

func (m *Mailer) SetFrom(name string, from string) error {
	if err := m.c.Mail(from); err != nil {
		return err
	}
	m.fromName = name
	m.fromMail = from
	return nil
}

func (m *Mailer) AddTo(name string, to string) error {
	if err := m.c.Rcpt(to); err != nil {
		return err
	}

	m.toNames = append(m.toNames, name)
	m.toEmails = append(m.toEmails, to)
	return nil
}

func (m *Mailer) Send(subject string, body string) error {
	wc, err := m.c.Data()
	if err != nil {
		return err
	}

	write := func(f string, a ...any) {
		if err != nil {
			return
		}

		_, err = wc.Write([]byte(fmt.Sprintf(f+"\r\n", a...)))
	}

	write("Subject: %s", mime.BEncoding.Encode("utf-8", subject))

	write("From: %s <%s>", mime.BEncoding.Encode("utf-8", m.fromName), m.fromMail)

	for i := 0; i < len(m.toNames); i++ {
		write("To: %s <%s>", mime.BEncoding.Encode("utf-8", m.toNames[i]), m.toEmails[i])
	}

	write("Content-Type: text/html; charset=utf-8")
	write("")

	write("%s", body)

	if err != nil {
		return err
	}

	if err = wc.Close(); err != nil {
		return err
	}

	return nil
}

func (m *Mailer) Quit() error {
	m.c.Quit()
	m.c.Close()
	m.conn.Close()
	return nil
}

type MailerConfig struct {
	server             string
	username, password string
}

func NewMailerConfig(server string, username, password string) MailerConfig {
	return MailerConfig{
		server:   server,
		username: username,
		password: password,
	}
}

type User struct {
	Name    string
	Address string
}

func (c *MailerConfig) Send(subject, body string, fromName string, tos []User) error {
	mc, err := DialTLS(c.server)
	if err != nil {
		return err
	}
	defer mc.Quit()
	if err := mc.Auth(c.username, c.password); err != nil {
		return err
	}
	if err := mc.SetFrom(fromName, c.username); err != nil {
		return err
	}
	for _, to := range tos {
		if err := mc.AddTo(to.Name, to.Address); err != nil {
			return err
		}
	}
	if err := mc.Send(subject, body); err != nil {
		return err
	}
	return nil
}

type MailerLogger struct {
	store logs.Logger

	mailer       MailerConfig
	pullInterval time.Duration
}

func NewMailerLogger(store logs.Logger, mailer MailerConfig) *MailerLogger {
	n := &MailerLogger{
		store:  store,
		mailer: mailer,
	}
	n.SetPullInterval(time.Second * 5)
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
	if d <= time.Millisecond*100 {
		d = time.Millisecond * 100
	}
	n.pullInterval = d
}

func (n *MailerLogger) Queue(subject, body string, fromName string, tos []User) {
	n.store.CreateLog(context.Background(), ty, sty, 1, _Message{
		FromName: fromName,
		Tos:      tos,
		Subject:  subject,
		Body:     body,
	})
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
				var te *textproto.Error
				if errors.As(err, &te) {
					if te.Code >= 500 {
						log.Println(`不可重试错误，将删除邮件。`)
						n.store.DeleteLog(ctx, l.ID)
						continue
					}
				}

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
