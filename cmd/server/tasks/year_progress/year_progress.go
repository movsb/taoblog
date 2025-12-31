package year_progress

import (
	"context"
	"fmt"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/calendar"
	"github.com/movsb/taoblog/service/modules/calendar/solar"
)

func isLeapYear(y int) bool {
	return y%400 == 0 || (y%4 == 0 && y%100 != 0)
}

// 返回当前进度百分比 [0-100]，-1为无效。
func calculate(t time.Time) int {
	var (
		yearTotal = utils.IIF(isLeapYear(t.Year()), 366, 365)
		yearDay   = t.YearDay()
	)

	switch yearDay {
	case 1:
		return 0
	case yearTotal:
		return 100
	default:
		// 只记录百分比发生跳变的这一天
		todayPercent := int(float32(yearDay) / float32(yearTotal) * 100)
		yesterdayPercent := int(float32(yearDay-1) / float32(yearTotal) * 100)
		if todayPercent == yesterdayPercent {
			return -1
		}
		return todayPercent
	}
}

var calKind = calendar.RegisterKind(func(e *calendar.Event) string {
	return e.Message
})

func schedule(cal *calendar.CalenderService) {
	cal.Remove(calKind, func(e *calendar.Event) bool {
		return true
	})

	now := time.Now()

	for i := -5; i <= +5; i++ {
		other := now.AddDate(0, 0, i)
		p := calculate(other)
		// 特别地：0% 也去除
		if p <= 0 {
			continue
		}

		st, et := solar.AllDay(other)

		cal.AddEvent(calKind, &calendar.Event{
			Message: fmt.Sprintf(`今年 %d%% 已过`, p),
			Start:   st,
			End:     et,
		})
	}
}

func New(ctx context.Context, cal *calendar.CalenderService) {
	schedule(cal)

	go utils.AtMiddleNight(ctx, func() {
		schedule(cal)
	})
}
