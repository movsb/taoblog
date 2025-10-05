package e2e_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
)

func TestCalendar(t *testing.T) {
	r := Serve(t.Context())

	_ = utils.Must1(r.client.Blog.CreatePost(r.user1, &proto.Post{
		Source: `# R1
~~~reminder
title: t1
dates:
  start: 2025-10-05
~~~
`,
		Status: models.PostStatusPublic,
	}))

	p2 := utils.Must1(r.client.Blog.CreatePost(r.user2, &proto.Post{
		Source: `# R2
~~~reminder
title: t2
dates:
  start: 2025-10-05
~~~
`,
		Status: models.PostStatusPrivate,
	}))

	refresh := func() {
		utils.Must1(r.client.Management.SetConfig(r.admin, &proto.SetConfigRequest{
			Path: `runtime.reminders.refresh_now`,
			Yaml: `1`,
		}))
	}

	refresh()
	time.Sleep(time.Second)

	const calPrefix = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//TaoBlog//Golang ICS Library
METHOD:PUBLISH
TIMEZONE-ID:Asia/Shanghai
X-WR-CALNAME:未命名`
	const calSuffix = `END:VCALENDAR`

	reRemoveUID := regexp.MustCompile(`(?m:^UID:.*\n)`)
	reLastModified := regexp.MustCompile(`(?m:^LAST-MODIFIED:.*\n)`)

	expect := func(ctx context.Context, calendar string) {
		settings := utils.Must1(r.client.Blog.GetUserSettings(ctx, &proto.GetUserSettingsRequest{}))
		rsp := utils.Must1(http.Get(settings.CalendarUrl))
		all := utils.Must1(io.ReadAll(rsp.Body))

		all = bytes.ReplaceAll(all, []byte("\r\n"), []byte("\n"))
		all = reRemoveUID.ReplaceAll(all, nil)
		all = reLastModified.ReplaceAll(all, nil)

		full := calPrefix + calendar + calSuffix
		if strings.TrimSpace(string(all)) != strings.TrimSpace(full) {
			_, file, line, _ := runtime.Caller(1)
			t.Fatalf("not equal: %s:%d\n%s\n\n%s", file, line, string(all), full)
		}
	}

	expect(r.user1, `
BEGIN:VEVENT
SUMMARY:t1
DTSTART;VALUE=DATE:20251005
DTEND;VALUE=DATE:20251006
END:VEVENT
`)

	expect(r.user2, `
BEGIN:VEVENT
SUMMARY:t2
DTSTART;VALUE=DATE:20251005
DTEND;VALUE=DATE:20251006
END:VEVENT
`)

	expect(r.admin, `
BEGIN:VEVENT
SUMMARY:t1
DTSTART;VALUE=DATE:20251005
DTEND;VALUE=DATE:20251006
END:VEVENT
BEGIN:VEVENT
SUMMARY:t2
DTSTART;VALUE=DATE:20251005
DTEND;VALUE=DATE:20251006
END:VEVENT
`)

	// 把 p2 分享给 p1 后日历也会分享
	utils.Must1(r.client.Blog.SetPostStatus(r.admin, &proto.SetPostStatusRequest{
		Id:     p2.Id,
		Status: models.PostStatusPartial,
	}))
	utils.Must1(r.client.Blog.SetPostACL(r.admin, &proto.SetPostACLRequest{
		PostId: p2.Id,
		Users: map[int32]*proto.UserPerm{
			int32(r.user1ID): {
				Perms: []proto.Perm{
					proto.Perm_PermRead,
				},
			},
		},
	}))

	refresh()
	time.Sleep(time.Second)

	expect(r.user1, `
BEGIN:VEVENT
SUMMARY:t1
DTSTART;VALUE=DATE:20251005
DTEND;VALUE=DATE:20251006
END:VEVENT
BEGIN:VEVENT
SUMMARY:t2
DTSTART;VALUE=DATE:20251005
DTEND;VALUE=DATE:20251006
END:VEVENT
`)
	expect(r.user2, `
BEGIN:VEVENT
SUMMARY:t2
DTSTART;VALUE=DATE:20251005
DTEND;VALUE=DATE:20251006
END:VEVENT
`)

	expect(r.admin, `
BEGIN:VEVENT
SUMMARY:t1
DTSTART;VALUE=DATE:20251005
DTEND;VALUE=DATE:20251006
END:VEVENT
BEGIN:VEVENT
SUMMARY:t2
DTSTART;VALUE=DATE:20251005
DTEND;VALUE=DATE:20251006
END:VEVENT
`)
}
