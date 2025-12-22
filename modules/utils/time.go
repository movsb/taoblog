package utils

import (
	"context"
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
