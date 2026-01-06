package utils

import (
	"context"
	"time"

	"github.com/xeonx/timeago"
)

type CurrentTimezoneGetter interface {
	GetCurrentTimezone() *time.Location
}

type LocalTimezoneGetter struct{}

func (LocalTimezoneGetter) GetCurrentTimezone() *time.Location {
	return time.Local
}

// 每天凌晨时执行函数。
// 函数不会主动返回，除非 ctx 完成。
// fn 在当前线程内执行。
func AtMiddleNight(ctx context.Context, fn func()) {
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-time.After(time.Second * 45):
			if now.Hour() == 0 && now.Minute() == 0 {
				fn()
				select {
				case <-ctx.Done():
					return
				case <-time.After(time.Hour * 2):
					// 担心时区变化，睡两个小时看看。
				}
			}
		}
	}
}

var _friendlyChineseDateFormat = timeago.Config{
	PastPrefix:   "",
	PastSuffix:   "前",
	FuturePrefix: "于",
	FutureSuffix: "",

	Periods: []timeago.FormatPeriod{
		{D: time.Second, One: "1秒", Many: "%d秒"},
		{D: time.Minute, One: "1分钟", Many: "%d分钟"},
		{D: time.Hour, One: "1小时", Many: "%d小时"},
		{D: timeago.Day, One: "1天", Many: "%d天"},
		{D: timeago.Month, One: "1月", Many: "%d月"},
		{D: timeago.Year, One: "1年", Many: "%d年"},
	},

	Zero: "1秒",

	Max:           73 * time.Hour,
	DefaultLayout: "2006-01-02",
}

func RelativeDateFrom(t time.Time, from time.Time) string {
	return _friendlyChineseDateFormat.FormatReference(t, from)
}

func RelativeDate(t time.Time) string {
	return _friendlyChineseDateFormat.FormatReference(t, time.Now())
}
