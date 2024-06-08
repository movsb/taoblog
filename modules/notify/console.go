package notify

import (
	"log"
)

type ConsoleNotifier struct {
}

func (n *ConsoleNotifier) InstantNotify(title, message string) {
	log.Println(title, message)
}

func NewConsoleNotify() InstantNotifier {
	return &ConsoleNotifier{}
}
