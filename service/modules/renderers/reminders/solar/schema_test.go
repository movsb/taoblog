package solar

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestDaysPassed(t *testing.T) {
	// 现在时间变化，事件时间不变。
	today := time.Date(2025, 8, 28, 23, 45, 0, 0, time.Local)

	// 现在是前一天
	yesterday := time.Date(2025, 8, 27, 0, 0, 0, 0, time.Local)
	expect(t, yesterday, today, false, -1)
	expect(t, yesterday, today, true, -1)

	// 现在是当天
	today2 := time.Date(2025, 8, 28, 0, 0, 0, 0, time.Local)
	expect(t, today2, today, false, 1)
	expect(t, today2, today, true, 0)

	// 现在是后一天
	tomorrow := time.Date(2025, 8, 29, 0, 0, 0, 0, time.Local)
	expect(t, tomorrow, today, false, 2)
	expect(t, tomorrow, today, true, 1)
}

func expect(t *testing.T, now, at time.Time, exclusive bool, n int) {
	if DaysPassed(now, at, exclusive) != n {
		_, file, line, _ := runtime.Caller(1)
		t.Fatal(`计算错误:`, now, at, exclusive, n, fmt.Sprintf(`%s:%d`, file, line))
	}
}
