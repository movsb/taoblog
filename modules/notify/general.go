package notify

import (
	"log"

	"github.com/movsb/pkg/notify"
)

type Notifier interface {
	Notify(title, message string) error
}

type _ChanifyNotify struct {
	chanify *notify.Chanify
}

func (n *_ChanifyNotify) Notify(title, message string) error {
	return n.chanify.Send(title, message, true)
}

func NewChanifyNotify(token string) Notifier {
	return &_ChanifyNotify{
		chanify: notify.NewOfficialChanify(token),
	}
}

type _ConsoleNotify struct{}

func (n *_ConsoleNotify) Notify(title, message string) error {
	log.Println(title, message)
	return nil
}

func NewConsoleNotify() Notifier {
	return &_ConsoleNotify{}
}
