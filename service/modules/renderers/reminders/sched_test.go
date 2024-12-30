package reminders_test

import (
	"log"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
)

func TestScheduler(t *testing.T) {
	// MacOS: date -j -v+99d -f %y%m%d 141224
	clock := clockwork.NewFakeClockAt(time.Date(2014, time.December, 24, 0, 0, 0, 0, time.Local))
	s := reminders.NewScheduler(clock)
	utils.Must(s.AddReminder(1,
		&reminders.Reminder{
			Title: `测试1`,
			Dates: reminders.ReminderDates{
				Start: utils.Must1(reminders.NewUserDateFromString(`2014-12-24`)),
			},
			Remind: reminders.ReminderRemind{
				Days: []int{100, 520, 1314},
			},
		},
		func() {
			log.Println(`调度测试1`)
		},
	))
}
