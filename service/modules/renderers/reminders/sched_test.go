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
UID:post_id:1\,job_id:1419350400
SUMMARY:测试第2天（包含当天）
DTSTAMP:20141223T160000Z
DTSTART;VALUE=DATE:20141224
DTEND;VALUE=DATE:20141225
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1419436800
SUMMARY:测试第2天（包含当天） 已经 2 天了
DTSTAMP:20141224T160000Z
DTSTART;VALUE=DATE:20141225
DTEND;VALUE=DATE:20141226
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
UID:post_id:1\,job_id:1419350400
SUMMARY:测试第2天（不包含当天）
DTSTAMP:20141223T160000Z
DTSTART;VALUE=DATE:20141224
DTEND;VALUE=DATE:20141225
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1419523200
SUMMARY:测试第2天（不包含当天） 已经 2 天了
DTSTAMP:20141225T160000Z
DTSTART;VALUE=DATE:20141226
DTEND;VALUE=DATE:20141227
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
UID:post_id:1\,job_id:1414684800
SUMMARY:测试月份
DTSTAMP:20141030T160000Z
DTSTART;VALUE=DATE:20141031
DTEND;VALUE=DATE:20141101
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1417276800
SUMMARY:测试月份 已经 1 个月了
DTSTAMP:20141129T160000Z
DTSTART;VALUE=DATE:20141130
DTEND;VALUE=DATE:20141201
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1419955200
SUMMARY:测试月份 已经 2 个月了
DTSTAMP:20141230T160000Z
DTSTART;VALUE=DATE:20141231
DTEND;VALUE=DATE:20150101
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1425052800
SUMMARY:测试月份 已经 4 个月了
DTSTAMP:20150227T160000Z
DTSTART;VALUE=DATE:20150228
DTEND;VALUE=DATE:20150301
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1427731200
SUMMARY:测试月份 已经 5 个月了
DTSTAMP:20150330T160000Z
DTSTART;VALUE=DATE:20150331
DTEND;VALUE=DATE:20150401
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
UID:post_id:1\,job_id:1456675200
SUMMARY:测试年份
DTSTAMP:20160228T160000Z
DTSTART;VALUE=DATE:20160229
DTEND;VALUE=DATE:20160301
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1488211200
SUMMARY:测试年份 已经 1 年了
DTSTAMP:20170227T160000Z
DTSTART;VALUE=DATE:20170228
DTEND;VALUE=DATE:20170301
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1582905600
SUMMARY:测试年份 已经 4 年了
DTSTAMP:20200228T160000Z
DTSTART;VALUE=DATE:20200229
DTEND;VALUE=DATE:20200301
END:VEVENT
BEGIN:VEVENT
UID:post_id:1\,job_id:1614441600
SUMMARY:测试年份 已经 5 年了
DTSTAMP:20210227T160000Z
DTSTART;VALUE=DATE:20210228
DTEND;VALUE=DATE:20210301
END:VEVENT
`,
		},
	}

	now := time.Date(2002, time.July, 3, 1, 2, 3, 0, time.Local)

	for _, test := range tests {
		sched := reminders.NewScheduler()
		r := utils.Must1(reminders.ParseReminder([]byte(test.Reminder)))
		utils.Must(sched.AddReminder(1, r))
		cal := reminders.NewCalendarService(`Cal`, sched)
		buf := bytes.NewBuffer(nil)
		utils.Must(cal.Marshal(now, buf))
		got := strings.ReplaceAll(strings.TrimSpace(buf.String()), "\r\n", "\n")
		want := strings.TrimSpace(test.Calendar)
		fullWant := calPrefix + want + calSuffix
		if got != fullWant {
			t1, _ := os.CreateTemp("", "got-*")
			t1.WriteString(got)
			t2, _ := os.CreateTemp("", "want-*")
			t2.WriteString(fullWant)
			t.Fatalf("输出不相等：\n%s\ngot:  %s\nwant: %s", test.Reminder, t1.Name(), t2.Name())
		}
	}
}
