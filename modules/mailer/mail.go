package mailer

import (
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
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
