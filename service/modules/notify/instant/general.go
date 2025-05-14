package instant

import (
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
