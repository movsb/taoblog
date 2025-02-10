package reminders

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jonboulle/clockwork"
	"github.com/movsb/taoblog/modules/utils"
)

type Scheduler struct {
	backend gocron.Scheduler
	clock   clockwork.Clock
}

type SchedulerOption func(s *Scheduler)

func WithFakeClock(clock clockwork.Clock) SchedulerOption {
	return func(s *Scheduler) {
		s.clock = clock
	}
}

func NewScheduler(options ...SchedulerOption) *Scheduler {
	sched := &Scheduler{}

	for _, opt := range options {
		opt(sched)
	}

	if sched.clock == nil {
		sched.clock = clockwork.NewRealClock()
	}

	backendOptions := []gocron.SchedulerOption{}
	if sched.clock != nil {
		backendOptions = append(backendOptions, gocron.WithClock(sched.clock))
	}

	sched.backend = utils.Must1(gocron.NewScheduler(backendOptions...))
	sched.backend.Start()

	return sched
}

func (s *Scheduler) AddReminder(r *Reminder, remind func(now time.Time, message string)) error {
	now := s.clock.Now()

	for _, day := range r.Remind.Days {
		if day == 1 {
			return fmt.Errorf(`提醒天数不能为第 1 天`)
		}
		t := time.Time(r.Dates.Start).AddDate(0, 0, utils.IIF(r.Exclusive, day, day-1))
		if t.Before(now) {
			continue
		}

		j, err := s.backend.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(t)),
			gocron.NewTask(remind, t, fmt.Sprintf(`%s 已经 %d 天了`, r.Title, day)),
		)
		if err != nil {
			log.Println(j, err)
			return err
		}
	}

	return nil
}
