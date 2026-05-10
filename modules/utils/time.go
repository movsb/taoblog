package utils

import (
	"context"
	"fmt"
	"time"
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

func RelativeDateFrom(t time.Time, from time.Time) string {
	diff := from.Sub(t)
	// 未来的时间，直接格式化。
	if diff < 0 {
		return t.Format(time.DateOnly)
	}
	switch {
	case diff < time.Minute:
		return "刚刚"
	case diff < time.Hour:
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	case diff < time.Hour*48 && t.Day() == from.Day():
		return fmt.Sprintf("%d小时前", int(diff.Hours()))
	}

	tt := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
	ff := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	diff = ff.Sub(tt)

	days := int(diff / time.Hour / 24)
	switch days {
	case 1:
		return "昨天"
	case 2:
		return "前天"
	}
	if days <= 30 {
		return fmt.Sprintf("%d天前", days)
	}
	if tt.Year() == ff.Year() {
		return t.Format(`01月02日`)
	}
	return t.Format(`2006年01月02日`)
}

func RelativeDate(t time.Time) string {
	return RelativeDateFrom(t, time.Now())
}
