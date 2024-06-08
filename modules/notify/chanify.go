package notify

import "github.com/movsb/pkg/notify"

type InstantNotifier interface {
	InstantNotify(title, message string)
}

type ChanifyInstantNotify struct {
	chanify *notify.Chanify
}

func (n *ChanifyInstantNotify) InstantNotify(title, message string) {
	n.chanify.Send(title, message, true)
}

func NewChanifyInstantNotify(token string) InstantNotifier {
	return &ChanifyInstantNotify{
		chanify: notify.NewOfficialChanify(token),
	}
}
