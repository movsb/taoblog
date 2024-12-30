package reminders

import (
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jonboulle/clockwork"
	"github.com/movsb/taoblog/modules/utils"
	"gopkg.in/yaml.v2"
)

type Scheduler struct {
	backend gocron.Scheduler
	clock   clockwork.Clock
}

func NewScheduler(clock clockwork.Clock) *Scheduler {
	sched := &Scheduler{
		backend: utils.Must1(gocron.NewScheduler(
			gocron.WithClock(clock),
		)),
		clock: clock,
	}
	sched.backend.Start()
	return sched
}

func (s *Scheduler) AddReminder(pid int64, r *Reminder, remind func()) error {
	now := s.clock.Now()

	if len(r.Remind.Days) > 0 {
		var dayTimes []time.Time
		for _, day := range r.Remind.Days {
			if day == 1 {
				return fmt.Errorf(`提醒天数不能为第 1 天`)
			}
			if !r.Exclusive {
				day--
			}
			t := time.Time(r.Dates.Start).AddDate(0, 0, day)
			if t.Before(now) {
				continue
			}
			dayTimes = append(dayTimes, t)
		}
		j, err := s.backend.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartDateTimes(dayTimes...)),
			gocron.NewTask(remind),
			gocron.WithTags(
				fmt.Sprintf(`post_id:%d`, pid),
			),
		)
		if err != nil {
			log.Println(j, err)
			return err
		}
		var nexts []time.Time
		for i := 1; i <= 100; i++ {
			ns, err := j.NextRuns(i)
			if err != nil {
				return err
			}
			if ns[i-1].IsZero() {
				break
			}
			nexts = append(nexts, ns[i-1])
		}
		log.Printf("%s:\n%s\n", j.Name(), string(utils.Must1(yaml.Marshal(nexts))))
	}
	return nil
}

func (s *Scheduler) newJob() {}
