package calendar

import (
	"io"
	"net/http"
	"slices"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/movsb/taoblog/modules/version"
)

// 日历中的一条独立的记录。
type Event struct {
	id  string // 事件的唯一编号
	now time.Time

	Message string

	Start time.Time
	End   time.Time

	UserID int
	PostID int

	Tags map[string]any
}

// 如果只包含日期（不包含时间），则认为是全天事件。
func (e *Event) isAllDay() bool {
	return isAllDay(e.Start, e.End)
}

type Events []*Event

func AddHeaders(w http.ResponseWriter) {
	w.Header().Set(`Content-Type`, `text/calendar; charset=utf-8`)
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
}
