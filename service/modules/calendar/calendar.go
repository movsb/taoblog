package calendar

import (
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

// 公共日历服务。
//
// https://icalendar.org/validator.html
type CalenderService struct {
	// 设置日历的名字，比如网站名。
	now func() time.Time

	lock   sync.Mutex
	events []*Event
}

func NewCalendarService(now func() time.Time) *CalenderService {
	s := &CalenderService{
		now: now,
	}
	return s
}

type Kind int

var nextID = Kind(1)

const AnyKind = Kind(0)

var uniqueFuncs = map[Kind]func(e *Event) string{}

func RegisterKind(unique func(e *Event) string) Kind {
	n := nextID
	nextID++
	uniqueFuncs[n] = unique
	return n
}

func (s *CalenderService) AddEvent(kind Kind, e *Event) {
	s.lock.Lock()
	defer s.lock.Unlock()

	e.id = uuid.NewString()
	e.now = s.now()
	e.kind = kind
	s.events = append(s.events, e)
}

func (s *CalenderService) Each(kind Kind, callback func(e *Event)) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, e := range s.events {
		if kind == AnyKind || e.kind == kind {
			callback(e)
		}
	}
}

func (s *CalenderService) Filter(kind Kind, predicate func(e *Event) bool) Events {
	events := Events{}

	s.lock.Lock()
	defer s.lock.Unlock()

	for _, e := range s.events {
		if (kind == AnyKind || e.kind == kind) && predicate(e) {
			events = append(events, e)
		}
	}

	return events
}

// 移除所有满足 predicate 的事件。
func (s *CalenderService) Remove(kind Kind, predicate func(e *Event) bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.events = slices.DeleteFunc(s.events, func(e *Event) bool {
		return (kind == AnyKind || e.kind == kind) && predicate(e)
	})
}

func (s *CalenderService) Unique(events Events) Events {
	return events.Unique(func(e *Event) string {
		return fmt.Sprintf(`%d:%s`, e.kind, uniqueFuncs[e.kind](e))
	})
}
