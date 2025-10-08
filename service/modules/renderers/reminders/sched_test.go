package reminders_test

import (
	"bytes"
	"os"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/calendar"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
)

const calPrefix = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//TaoBlog//Golang ICS Library
METHOD:PUBLISH
TIMEZONE-ID:Asia/Shanghai
X-WR-CALNAME:Cal
`
const calSuffix = `
END:VCALENDAR`

var reRemoveUID = regexp.MustCompile(`(?m:^UID:.*\n)`)
var reLastModified = regexp.MustCompile(`(?m:^LAST-MODIFIED:.*\n)`)

func runCal(t *testing.T, cal *calendar.CalenderService, sched *reminders.Scheduler, reminder string, expect string) {
	sched.DeleteRemindersByPostID(1)

	r := utils.Must1(reminders.ParseReminder([]byte(reminder)))

	// debugger
	if r.Title == `每天` {
		r.Title += ``
	}

	utils.Must(sched.AddReminder(1, 1, r))

	all := cal.Filter(func(e *calendar.Event) bool { return true })
	buf := bytes.NewBuffer(nil)
	all.Marshal(`Cal`, buf)
	got := strings.ReplaceAll(strings.TrimSpace(buf.String()), "\r\n", "\n")
	got = reRemoveUID.ReplaceAllLiteralString(got, ``)
	got = reLastModified.ReplaceAllLiteralString(got, ``)

	want := strings.TrimSpace(expect)
	fullWant := calPrefix + want + calSuffix

	if got != fullWant {
		t1, _ := os.Create("/tmp/1")
		t1.WriteString(got)
		t2, _ := os.Create("/tmp/2")
		t2.WriteString(fullWant)
		t.Logf("输出不相等：\n%s\ngot:  %s\nwant: %s", reminder, t1.Name(), t2.Name())
		_, file, line, _ := runtime.Caller(1)
		t.Logf(`%s:%d`, file, line)
		t.Fatalf("输出不相等：\ngot: \n%s\n\nwant: %s", got, fullWant)
	}
}

func TestScheduler(t *testing.T) {
	tests := []struct {
		Reminder string
		Calendar string
	}{
		{
			Reminder: `
title: 测试第2天（包含当天）
dates:
  start: 2014-12-24
remind:
  days: [2]
`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:测试第2天（包含当天） 已经 2 天了
DTSTART;VALUE=DATE:20141225
DTEND;VALUE=DATE:20141226
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试第2天（包含当天）
DTSTART;VALUE=DATE:20141224
DTEND;VALUE=DATE:20141225
END:VEVENT
`,
		},
		{
			Reminder: `
title: 测试第2天（不包含当天）
dates:
  start: 2014-12-24
exclusive: true
remind:
  days: [2]
`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:测试第2天（不包含当天） 已经 2 天了
DTSTART;VALUE=DATE:20141226
DTEND;VALUE=DATE:20141227
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试第2天（不包含当天）
DTSTART;VALUE=DATE:20141224
DTEND;VALUE=DATE:20141225
END:VEVENT
`,
		},
		{
			Reminder: `
title: 测试周数
dates:
  start: 2014-10-31
remind:
  weeks: [1,5]
`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:测试周数 已经 1 周了
DTSTART;VALUE=DATE:20141107
DTEND;VALUE=DATE:20141108
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试周数 已经 5 周了
DTSTART;VALUE=DATE:20141205
DTEND;VALUE=DATE:20141206
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试周数
DTSTART;VALUE=DATE:20141031
DTEND;VALUE=DATE:20141101
END:VEVENT
`,
		},
		{
			Reminder: `
title: 测试月份
dates:
  start: 2014-10-31
remind:
  months: [1,2,4,5]
`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:测试月份 已经 1 个月了
DTSTART;VALUE=DATE:20141130
DTEND;VALUE=DATE:20141201
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试月份 已经 2 个月了
DTSTART;VALUE=DATE:20141231
DTEND;VALUE=DATE:20150101
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试月份 已经 4 个月了
DTSTART;VALUE=DATE:20150228
DTEND;VALUE=DATE:20150301
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试月份 已经 5 个月了
DTSTART;VALUE=DATE:20150331
DTEND;VALUE=DATE:20150401
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试月份
DTSTART;VALUE=DATE:20141031
DTEND;VALUE=DATE:20141101
END:VEVENT
`,
		},
		{
			Reminder: `
title: 测试年份
dates:
  start: 2016-02-29
remind:
  years: [1,4,5]
`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:测试年份 已经 1 年了
DTSTART;VALUE=DATE:20170228
DTEND;VALUE=DATE:20170301
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试年份 已经 4 年了
DTSTART;VALUE=DATE:20200229
DTEND;VALUE=DATE:20200301
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试年份 已经 5 年了
DTSTART;VALUE=DATE:20210228
DTEND;VALUE=DATE:20210301
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试年份
DTSTART;VALUE=DATE:20160229
DTEND;VALUE=DATE:20160301
END:VEVENT
`,
		},
		{
			Reminder: `
title: 测试每天
dates:
  start: 2002-07-01
remind:
  daily: true
`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:测试每天 已经 3 天了
DTSTART;VALUE=DATE:20020703
DTEND;VALUE=DATE:20020704
END:VEVENT
`,
		},
		{
			Reminder: `
title: 测试前2天
dates:
  start: 2025-03-03
remind:
  days: [-2]
`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:测试前2天
DTSTART;VALUE=DATE:20250303
DTEND;VALUE=DATE:20250305
END:VEVENT
`,
		},
		{
			Reminder: `
title: 测试前2周
dates:
  start: 2025-03-03
remind:
  weeks: [-2]
`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:测试前2周
DTSTART;VALUE=DATE:20250303
DTEND;VALUE=DATE:20250304
END:VEVENT
BEGIN:VEVENT
SUMMARY:测试前2周
DTSTART;VALUE=DATE:20250310
DTEND;VALUE=DATE:20250311
END:VEVENT
`,
		},
		{
			Reminder: `
title: 前一周带时间
dates:
  start: 2025-03-03 13:30
  end: 2025-03-03 17:10
remind:
  weeks: [-1]
`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:前一周带时间
DTSTART:20250303T053000Z
DTEND:20250303T091000Z
END:VEVENT
`,
		},
		{
			Reminder: `
title: 武胜 → 重庆（D156）
dates:
  start: 2025-03-07 19:20
  end: 2025-03-07 20:19

`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:武胜 → 重庆（D156）
DTSTART:20250307T112000Z
DTEND:20250307T121900Z
END:VEVENT`,
		},
		{
			Reminder: `
title: 每天
dates:
  start: 2025-10-06
remind:
  every: [1d]`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:每天 已经 1 天了
DTSTART;VALUE=DATE:20251006
DTEND;VALUE=DATE:20251007
END:VEVENT`,
		},
		{
			Reminder: `
title: 每天（不包含当天）
dates:
  start: 2002-07-03
exclusive: true
remind:
  every: [1d]`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:每天（不包含当天） 已经 1 天了
DTSTART;VALUE=DATE:20020704
DTEND;VALUE=DATE:20020705
END:VEVENT`,
		},
		{
			Reminder: `
title: 每周月年
dates:
  start: 2025-10-06
remind:
  every: [1d,2w,3m,4y]`,
			Calendar: `
BEGIN:VEVENT
SUMMARY:每周月年 已经 1 天了
DTSTART;VALUE=DATE:20251006
DTEND;VALUE=DATE:20251007
END:VEVENT
BEGIN:VEVENT
SUMMARY:每周月年 已经 2 周了
DTSTART;VALUE=DATE:20251020
DTEND;VALUE=DATE:20251021
END:VEVENT
BEGIN:VEVENT
SUMMARY:每周月年 已经 3 个月了
DTSTART;VALUE=DATE:20260106
DTEND;VALUE=DATE:20260107
END:VEVENT
BEGIN:VEVENT
SUMMARY:每周月年 已经 4 年了
DTSTART;VALUE=DATE:20291006
DTEND;VALUE=DATE:20291007
END:VEVENT`,
		},
	}

	fixed := time.FixedZone(`fixed`, 8*60*60)
	now := time.Date(2002, time.July, 3, 1, 2, 3, 0, fixed)

	for _, test := range tests {
		cal := calendar.NewCalendarService(func() time.Time { return now })
		sched := reminders.NewScheduler(cal, func() time.Time { return now })
		runCal(t, cal, sched, test.Reminder, test.Calendar)
	}
}

func TestUpdateDaily(t *testing.T) {
	fixed := time.FixedZone(`fixed`, 8*60*60)
	var now time.Time

	cal := calendar.NewCalendarService(func() time.Time { return now })
	sched := reminders.NewScheduler(cal, func() time.Time { return now })

	reminder := `
title: t1
dates:
  start: 2002-07-04
remind:
  daily: true
`

	// 2002-07-03
	now = time.Date(2002, time.July, 3, 1, 2, 3, 0, fixed)

	runCal(t, cal, sched, reminder, `
BEGIN:VEVENT
SUMMARY:t1 已经 -1 天了
DTSTART;VALUE=DATE:20020703
DTEND;VALUE=DATE:20020704
END:VEVENT
`)

	// 2002-07-10
	now = now.AddDate(0, 0, 7)

	runCal(t, cal, sched, reminder, `
BEGIN:VEVENT
SUMMARY:t1 已经 7 天了
DTSTART;VALUE=DATE:20020710
DTEND;VALUE=DATE:20020711
END:VEVENT
`)
}

func TestEvery(t *testing.T) {
	fixed := time.FixedZone(`fixed`, 8*60*60)
	var now time.Time

	cal := calendar.NewCalendarService(func() time.Time { return now })
	sched := reminders.NewScheduler(cal, func() time.Time { return now })

	reminder := `
title: 给车充电
dates:
  start: 2025-04-14
remind:
  every: [1w]
`

	// 2025-10-06
	now = time.Date(2025, time.October, 6, 14, 30, 0, 0, fixed)

	runCal(t, cal, sched, reminder, `
BEGIN:VEVENT
SUMMARY:给车充电 已经 25 周了
DTSTART;VALUE=DATE:20251006
DTEND;VALUE=DATE:20251007
END:VEVENT
BEGIN:VEVENT
SUMMARY:给车充电
DTSTART;VALUE=DATE:20250414
DTEND;VALUE=DATE:20250415
END:VEVENT
`)
}
