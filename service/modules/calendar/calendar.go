package calendar

import (
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

func (s *CalenderService) AddEvent(e *Event) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := e.Tags[`uuid`]; !ok {
		panic(`need uuid`)
	}

	e.id = uuid.NewString()
	e.now = s.now()
	s.events = append(s.events, e)
}

func (s *CalenderService) Each(callback func(e *Event)) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, e := range s.events {
		callback(e)
	}
}

func (s *CalenderService) Filter(predicate func(e *Event) bool) Events {
	events := Events{}

	s.lock.Lock()
	defer s.lock.Unlock()

	for _, e := range s.events {
		if predicate(e) {
			events = append(events, e)
		}
	}

	return events
}

// 移除所有满足 predicate 的事件。
func (s *CalenderService) Remove(predicate func(e *Event) bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.events = slices.DeleteFunc(s.events, predicate)
}
