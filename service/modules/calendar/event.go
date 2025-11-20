package calendar

import (
	"io"
	"net/http"
	"slices"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/service/modules/calendar/solar"
)

// 日历中的一条独立的记录。
type Event struct {
	id     string // 事件的唯一编号
	now    time.Time
	kind   Kind // 种类
	unique func() string

	Message string // 事件标题。

	Start time.Time
	End   time.Time

	// 用户ID不一定是文章作者，可能来自分享。
	UserID int
	PostID int

	URL         string // 链接地址。
	Description string // 详细描述信息。

	Tags map[string]any
}

func (e *Event) Unique() string {
	return e.unique()
}

// 如果只包含日期（不包含时间），则认为是全天事件。
func (e *Event) isAllDay() bool {
	return solar.IsAllDay(e.Start, e.End)
}

type Events []*Event

func AddHeaders(w http.ResponseWriter) {
	w.Header().Set(`Content-Type`, `text/calendar; charset=utf-8`)
}

// 同一份日历会因为共享的原因存在于两个用户下，并且由于管理员
// 可以查看所有用户日历的原因，会存在重复，需要去重。
func (es Events) Unique(id func(e *Event) string) (out Events) {
	m := map[string]struct{}{}
	for _, e := range es {
		if _, ok := m[id(e)]; !ok {
			m[id(e)] = struct{}{}
			out = append(out, e)
		}
	}
	return
}

func (es Events) Marshal(name string, w io.Writer) {
	cal := ics.NewCalendarFor(version.Name)
	cal.SetMethod(ics.MethodPublish)

	defer cal.SerializeTo(w, ics.WithNewLine("\r\n"))

	if len(es) <= 0 {
		return
	}

	// 此列表的最新时间。
	newest := slices.MaxFunc(es, func(a, b *Event) int {
		return int(a.now.Unix() - b.now.Unix())
	})
	cal.SetLastModified(newest.now)

	// TODO 写死了
	cal.SetTimezoneId(`Asia/Shanghai`)
	cal.SetXWRCalName(name)

	for _, e := range es {
		es.single(cal, e)
	}
}

func (es Events) single(cal *ics.Calendar, event *Event) {
	e := cal.AddEvent(event.id)
	e.SetSummary(event.Message)

	// 不是很清楚这个的作用。
	// e.SetDtStampTime(event.Start)

	if event.isAllDay() {
		e.SetAllDayStartAt(event.Start)
		e.SetAllDayEndAt(event.End)
	} else {
		e.SetStartAt(event.Start)
		e.SetEndAt(event.End)
	}

	if event.URL != `` {
		e.SetURL(event.URL)
	}
	if event.Description != `` {
		e.SetDescription(event.Description)
	}
}
