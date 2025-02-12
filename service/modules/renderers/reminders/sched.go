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

	createJob := func(t time.Time, message string) error {
		j, err := s.backend.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(t)),
			gocron.NewTask(remind, t, message),
			gocron.WithTags(r.tags...),
		)
		if err != nil {
			log.Println(r, err)
			return err
		}
		_ = j
		return nil
	}

	for _, day := range r.Remind.Days {
		if day == 1 {
			return fmt.Errorf(`提醒天数不能为第 1 天`)
		}
		t := time.Time(r.Dates.Start).AddDate(0, 0, utils.IIF(r.Exclusive, day, day-1))
		if t.Before(now) {
			continue
		}

		if err := createJob(t, fmt.Sprintf(`%s 已经 %d 天了`, r.Title, day)); err != nil {
			return err
		}
	}

	for _, month := range r.Remind.Months {
		if month < 1 {
			return fmt.Errorf(`提醒月份不能小于 1 个月`)
		}

		d1 := time.Time(r.Dates.Start)
		d2 := d1.AddDate(0, month, 0)

		// 注意 AddDate：
		//
		// 2014-10-31 +1 个月，期待：2014-11-30，但实际会是 2014-12-01 号。
		// 2014-12-31 +2 个月，期待：2015-02-28，但实际会是 2015-03-03 号。
		//
		// 实际结果均与目前的设计有违，手动往前调整到上个月最后一天。
		//
		// 注意，12月到1月会 round
		expect := int(d1.Month()) + month
		if expect > 12 {
			expect -= 12
		}
		for expect != int(d2.Month()) {
			d2 = d2.AddDate(0, 0, -1)
		}

		if d2.Before(now) {
			continue
		}

		if err := createJob(d2, fmt.Sprintf(`%s 已经 %d 个月了`, r.Title, month)); err != nil {
			return err
		}
	}

	for _, year := range r.Remind.Years {
		if year < 1 {
			return fmt.Errorf(`提醒年份不能小于 1 年`)
		}

		d1 := time.Time(r.Dates.Start)
		d2 := d1.AddDate(year, 0, 0)

		// 同上面月份的注意事项
		expect := int(d1.Month())
		for expect != int(d2.Month()) {
			d2 = d2.AddDate(0, 0, -1)
		}

		if d2.Before(now) {
			continue
		}

		if err := createJob(d2, fmt.Sprintf(`%s 已经 %d 年了`, r.Title, year)); err != nil {
			return err
		}
	}

	log.Println(`提醒：添加任务：`, r.Title)

	return nil
}

// 根据标签删除提醒。
func (s *Scheduler) DeleteRemindersByTags(tags ...string) {
	s.backend.RemoveByTags(tags...)
}
