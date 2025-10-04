package calendar_test

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/movsb/taoblog/service/modules/calendar"
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
	if calendar.DaysPassed(now, at, exclusive) != n {
		_, file, line, _ := runtime.Caller(1)
		t.Fatal(`计算错误:`, now, at, exclusive, n, fmt.Sprintf(`%s:%d`, file, line))
	}
}

func parse(layout, t string) time.Time {
	t2, err := time.ParseInLocation(layout, t, time.Local)
	if err != nil {
		panic(err)
	}
	return t2
}

func TestDaily(t *testing.T) {
	// {
	// 	layout := `2006-01-02`
	// 	tcs := [][5]string{
	// 		{
	// 			`2002-07-03`,
	// 			`2002-07-03`,
	// 			`2002-07-04`,
	// 			`2002-07-03`,
	// 			`2002-07-04`,
	// 		},
	// 		{
	// 			`2002-07-04`,
	// 			`2002-07-03`,
	// 			`2002-07-03`,
	// 			`2002-07-04`,
	// 			`2002-07-05`,
	// 		},
	// 	}
	// 	for _, tc := range tcs {
	// 		now := parse(layout, tc[0])
	// 		s, e := parse(layout, tc[1]), parse(layout, tc[2])
	// 		es := parse(layout, tc[3])
	// 		ee := parse(layout, tc[4])

	// 		gotES, gotEE := calendar.Daily(now, s, e)
	// 		if !es.Equal(gotES) {
	// 			panic(`not equal`)
	// 		}
	// 		if !ee.Equal(gotEE) {
	// 			panic(`not equal`)
	// 		}
	// 	}
	// }
	{
		layout := `2006-01-02 15:04`
		tcs := [][5]string{
			{
				`2002-07-03 13:00`,
				`2002-07-03 13:00`,
				`2002-07-03 14:00`,
				`2002-07-03 13:00`,
				`2002-07-03 14:00`,
			},
			{
				`2002-07-04 18:00`,
				`2002-07-03 13:00`,
				`2002-07-03 14:00`,
				`2002-07-04 13:00`,
				`2002-07-04 14:00`,
			},
		}
		for _, tc := range tcs {
			now := parse(layout, tc[0])
			s, e := parse(layout, tc[1]), parse(layout, tc[2])
			es := parse(layout, tc[3])
			ee := parse(layout, tc[4])

			gotES, gotEE := calendar.Daily(now, s, e)
			if !es.Equal(gotES) {
				panic(`not equal`)
			}
			if !ee.Equal(gotEE) {
				panic(`not equal`)
			}
		}
	}
}

func TestFirstDays(t *testing.T) {
	{
		layout := `2006-01-02`
		start := `2002-07-03`
		end := `2002-07-04`
		times := calendar.FirstDays(parse(layout, start), parse(layout, end), 3)
		expect := [][2]time.Time{
			{
				parse(layout, `2002-07-03`),
				parse(layout, `2002-07-06`),
			},
		}
		if !reflect.DeepEqual(times, expect) {
			t.Errorf("not equal: \n%v\n%v", times, expect)
		}
	}
	{
		// 无效全天
		layout := `2006-01-02`
		start := `2002-07-03`
		end := `2002-07-05`
		times := calendar.FirstDays(parse(layout, start), parse(layout, end), 3)
		expect := [][2]time.Time{
			{
				parse(layout, `2002-07-03`),
				parse(layout, `2002-07-05`),
			},
		}
		if !reflect.DeepEqual(times, expect) {
			t.Errorf("not equal: \n%v\n%v", times, expect)
		}
	}
	{
		layout := `2006-01-02 15:04`
		start := `2002-07-03 13:00`
		end := `2002-07-03 14:00`
		times := calendar.FirstDays(parse(layout, start), parse(layout, end), 3)
		expect := [][2]time.Time{
			{
				parse(layout, `2002-07-03 13:00`),
				parse(layout, `2002-07-03 14:00`),
			},
			{
				parse(layout, `2002-07-04 13:00`),
				parse(layout, `2002-07-04 14:00`),
			},
			{
				parse(layout, `2002-07-05 13:00`),
				parse(layout, `2002-07-05 14:00`),
			},
		}
		if !reflect.DeepEqual(times, expect) {
			t.Errorf("not equal: \n%v\n%v", times, expect)
		}
	}
}

func TestFirstWeeks(t *testing.T) {
	{
		layout := `2006-01-02`
		start := `2025-10-04`
		times := calendar.FirstWeeks(parse(layout, start), parse(layout, start), 3)
		expect := [][2]time.Time{
			{
				parse(layout, `2025-10-04`),
				parse(layout, `2025-10-04`),
			},
			{
				parse(layout, `2025-10-11`),
				parse(layout, `2025-10-11`),
			},
			{
				parse(layout, `2025-10-18`),
				parse(layout, `2025-10-18`),
			},
		}
		if !reflect.DeepEqual(times, expect) {
			t.Errorf("not equal: \n%v\n%v", times, expect)
		}
	}
	{
		layout := `2006-01-02 15:04`
		start := `2025-10-04 08:00`
		end := `2025-10-04 12:00`
		times := calendar.FirstWeeks(parse(layout, start), parse(layout, end), 3)
		expect := [][2]time.Time{
			{
				parse(layout, `2025-10-04 08:00`),
				parse(layout, `2025-10-04 12:00`),
			},
			{
				parse(layout, `2025-10-11 08:00`),
				parse(layout, `2025-10-11 12:00`),
			},
			{
				parse(layout, `2025-10-18 08:00`),
				parse(layout, `2025-10-18 12:00`),
			},
		}
		if !reflect.DeepEqual(times, expect) {
			t.Errorf("not equal: \n%v\n%v", times, expect)
		}
	}
}
