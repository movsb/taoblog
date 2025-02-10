package reminders_test

import (
	"slices"
	"testing"
	"time"

	"github.com/Lofanmi/chinese-calendar-golang/calendar"
	"github.com/jonboulle/clockwork"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
)

func TestScheduler(t *testing.T) {
	// MacOS: date -j -v+99d -f %y%m%d 141224
	d := time.Date(2014, time.December, 24, 0, 0, 0, 0, time.Local)

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
				d.AddDate(0, 0, 1),
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
				d.AddDate(0, 0, 2),
			},
		},
		{
			Reminder: reminders.Reminder{
				Title: `测试1（包含当天）`,
				Dates: reminders.DateStart(`2014-12-24`),
				Remind: reminders.ReminderRemind{
					Days: []int{100, 520, 1314},
				},
			},
			Dates: []time.Time{
				d.AddDate(0, 0, 100-1),
				d.AddDate(0, 0, 520-1),
				d.AddDate(0, 0, 1314-1),
			},
		},
	}

	for _, tt := range tests {
		func() {
			f := clockwork.NewFakeClockAt(d)
			s := reminders.NewScheduler(reminders.WithFakeClock(f))
			var ts []time.Time
			utils.Must(s.AddReminder(&tt.Reminder, func(now time.Time, message string) {
				ts = append(ts, now)
			}))
			f.Advance(time.Hour * 24 * 365 * 100)
			time.Sleep(time.Second)
			slices.SortFunc(tt.Dates, func(a, b time.Time) int { return int(a.UnixNano() - b.UnixNano()) })
			slices.SortFunc(ts, func(a, b time.Time) int { return int(a.UnixNano() - b.UnixNano()) })
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
		}()
	}
}

func TestLunar(t *testing.T) {
	// 杨家大院建成日期：2005年3月初8 == 2005-4-16
	builtAt := calendar.ByLunar(2005, 3, 8, 0, 0, 0, false)
	y := builtAt.Solar.GetYear()
	m := builtAt.Solar.GetMonth()
	d := builtAt.Solar.GetDay()
	if !(y == 2005 && m == 4 && d == 16) {
		t.Fatal(`农历、阳历不相等。`)
	}
}
