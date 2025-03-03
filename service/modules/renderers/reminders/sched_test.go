package reminders_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
)

const calPrefix = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//TaoBlog//Golang ICS Library
METHOD:PUBLISH
LAST-MODIFIED:20020702T170203Z
TIMEZONE-ID:Asia/Shanghai
X-WR-CALNAME:Cal
`
const calSuffix = `
END:VCALENDAR`

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
UID:post_id:1\,job_id:1419436800\,title:286e1e8e
SUMMARY:测试第2天（包含当天） 已经 2 天了
DTSTAMP:20141224T160000Z
DTSTART;VALUE=DATE:20141225
DTEND;VALUE=DATE:20141226
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1419350400\,title:25804a8e
SUMMARY:测试第2天（包含当天）
DTSTAMP:20141223T160000Z
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
UID:post_id:1\,job_id:1419523200\,title:e68b4d8f
SUMMARY:测试第2天（不包含当天） 已经 2 天了
DTSTAMP:20141225T160000Z
DTSTART;VALUE=DATE:20141226
DTEND;VALUE=DATE:20141227
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1419350400\,title:56f84946
SUMMARY:测试第2天（不包含当天）
DTSTAMP:20141223T160000Z
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
UID:post_id:1\,job_id:1415289600\,title:69d3793d
SUMMARY:测试周数 已经 1 周了
DTSTAMP:20141106T160000Z
DTSTART;VALUE=DATE:20141107
DTEND;VALUE=DATE:20141108
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1417708800\,title:ed9977c7
SUMMARY:测试周数 已经 5 周了
DTSTAMP:20141204T160000Z
DTSTART;VALUE=DATE:20141205
DTEND;VALUE=DATE:20141206
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1414684800\,title:eb3f1ba2
SUMMARY:测试周数
DTSTAMP:20141030T160000Z
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
UID:post_id:1\,job_id:1417276800\,title:d2a41df1
SUMMARY:测试月份 已经 1 个月了
DTSTAMP:20141129T160000Z
DTSTART;VALUE=DATE:20141130
DTEND;VALUE=DATE:20141201
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1419955200\,title:4b467bf0
SUMMARY:测试月份 已经 2 个月了
DTSTAMP:20141230T160000Z
DTSTART;VALUE=DATE:20141231
DTEND;VALUE=DATE:20150101
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1425052800\,title:a3f3b1b3
SUMMARY:测试月份 已经 4 个月了
DTSTAMP:20150227T160000Z
DTSTART;VALUE=DATE:20150228
DTEND;VALUE=DATE:20150301
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1427731200\,title:627d6e73
SUMMARY:测试月份 已经 5 个月了
DTSTAMP:20150330T160000Z
DTSTART;VALUE=DATE:20150331
DTEND;VALUE=DATE:20150401
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1414684800\,title:433ba37c
SUMMARY:测试月份
DTSTAMP:20141030T160000Z
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
UID:post_id:1\,job_id:1488211200\,title:210640e5
SUMMARY:测试年份 已经 1 年了
DTSTAMP:20170227T160000Z
DTSTART;VALUE=DATE:20170228
DTEND;VALUE=DATE:20170301
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1582905600\,title:69e64e81
SUMMARY:测试年份 已经 4 年了
DTSTAMP:20200228T160000Z
DTSTART;VALUE=DATE:20200229
DTEND;VALUE=DATE:20200301
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1614441600\,title:a54c4e1f
SUMMARY:测试年份 已经 5 年了
DTSTAMP:20210227T160000Z
DTSTART;VALUE=DATE:20210228
DTEND;VALUE=DATE:20210301
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1456675200\,title:767336bf
SUMMARY:测试年份
DTSTAMP:20160228T160000Z
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
UID:post_id:1\,job_id:1025452800\,title:ee2fe947
SUMMARY:测试每天
DTSTAMP:20020630T160000Z
DTSTART;VALUE=DATE:20020701
DTEND;VALUE=DATE:20020702
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1025452800\,daily:true
SUMMARY:测试每天 已经 3 天了
DTSTAMP:20020630T160000Z
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
UID:post_id:1\,job_id:1740931200\,first_days:2
SUMMARY:测试前2天
DTSTAMP:20250302T160000Z
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
UID:post_id:1\,job_id:1740931200\,first_weeks:2
SUMMARY:测试前2周
DTSTAMP:20250302T160000Z
DTSTART;VALUE=DATE:20250303
DTEND;VALUE=DATE:20250305
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
UID:post_id:1\,job_id:1740979800\,first_weeks:1\,week:1
SUMMARY:前一周带时间
DTSTAMP:20250303T053000Z
DTSTART:20250303T053000Z
DTEND:20250303T091000Z
END:VEVENT
`,
		},
	}

	fixed := time.FixedZone(`fixed`, 8*60*60)
	now := time.Date(2002, time.July, 3, 1, 2, 3, 0, fixed)

	for _, test := range tests {
		sched := reminders.NewScheduler(reminders.WithNowFunc(func() time.Time { return now }))
		r := utils.Must1(reminders.ParseReminder([]byte(test.Reminder)))
		utils.Must(sched.AddReminder(1, r))
		cal := reminders.NewCalendarService(`Cal`, sched)
		buf := bytes.NewBuffer(nil)
		utils.Must(cal.Marshal(now, buf))
		got := strings.ReplaceAll(strings.TrimSpace(buf.String()), "\r\n", "\n")
		want := strings.TrimSpace(test.Calendar)
		fullWant := calPrefix + want + calSuffix
		if got != fullWant {
			t1, _ := os.Create("/tmp/1")
			t1.WriteString(got)
			t2, _ := os.Create("/tmp/2")
			t2.WriteString(fullWant)
			t.Logf("输出不相等：\n%s\ngot:  %s\nwant: %s", test.Reminder, t1.Name(), t2.Name())
			t.Fatalf("输出不相等：\n%s\ngot:  %s\nwant: %s", test.Reminder, got, fullWant)
		}
	}
}
