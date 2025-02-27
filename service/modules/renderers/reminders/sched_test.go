package reminders_test

import (
	"slices"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
)

func TestScheduler(t *testing.T) {
	date := func(y, m, d int) time.Time {
		return time.Date(y, time.Month(m), d, 0, 0, 0, 0, reminders.FixedZone)
	}

	var tests = []struct {
		Reminder reminders.Reminder
		Dates    []time.Time
	}{
		{
			Reminder: reminders.Reminder{
				Title:     `测试第2天（包含当天）`,
				Dates:     reminders.DateStart(`2014-12-24`),
				Exclusive: false,
				Remind: reminders.ReminderRemind{
					Days: []int{2},
				},
			},
			Dates: []time.Time{
				date(2014, 12, 25),
			},
		},
		{
			Reminder: reminders.Reminder{
				Title:     `测试第2天（不包含当天）`,
				Dates:     reminders.DateStart(`2014-12-24`),
				Exclusive: true,
				Remind: reminders.ReminderRemind{
					Days: []int{2},
				},
			},
			Dates: []time.Time{
				date(2014, 12, 26),
			},
		},
		{
			Reminder: reminders.Reminder{
				Title: `测试月份`,
				Dates: reminders.DateStart(`2014-10-31`),
				Remind: reminders.ReminderRemind{
					Months: []int{1, 2, 4, 5},
				},
			},
			Dates: []time.Time{
				date(2014, 11, 30),
				date(2014, 12, 31),
				date(2015, 2, 28),
				date(2015, 3, 31),
			},
		},
		{
			Reminder: reminders.Reminder{
				Title: `测试年份`,
				Dates: reminders.DateStart(`2016-02-29`),
				Remind: reminders.ReminderRemind{
					Years: []int{1, 4, 5},
				},
			},
			Dates: []time.Time{
				date(2017, 2, 28),
				date(2020, 2, 29),
				date(2021, 2, 28),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Reminder.Title, func(t *testing.T) {
			t.Parallel()
			d := time.Time(tt.Reminder.Dates.Start)
			f := clockwork.NewFakeClockAt(d)
			s := reminders.NewScheduler(reminders.WithFakeClock(f))
			var ts []time.Time
			utils.Must(s.AddReminder(1, &tt.Reminder, func(now time.Time, message string) {
				ts = append(ts, now)
			}))
			f.Advance(time.Hour * 24 * 365 * 100)
			time.Sleep(time.Second * 2)
			slices.SortFunc(tt.Dates, func(a, b time.Time) int {
				return int(a.UnixNano() - b.UnixNano())
			})
			slices.SortFunc(ts, func(a, b time.Time) int {
				return int(a.UnixNano() - b.UnixNano())
			})
			if slices.CompareFunc(tt.Dates, ts, func(a, b time.Time) int {
				y1, m1, d1 := a.Date()
				y2, m2, d2 := b.Date()
				if y1 == y2 && m1 == m2 && d1 == d2 {
					return 0
				}
				return -1
			}) != 0 {
				t.Errorf("%s: 不相等：\n期望：%v\n实际：%v", tt.Reminder.Title, tt.Dates, ts)
			}
		})
	}
}
