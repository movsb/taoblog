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
func AtMiddleNight(ctx context.Context, fn func()) {
	ticker := time.NewTicker(time.Second * 50)
	defer ticker.Stop()
	last := time.Now()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if (now.Hour() == 0 && now.Minute() == 0) || (last.Day() != now.Day()) {
				fn()
				last = now
			}
		}
	}
}
