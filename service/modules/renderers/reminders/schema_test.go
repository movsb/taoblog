package reminders_test

import (
	"testing"

	"github.com/Lofanmi/chinese-calendar-golang/calendar"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
)

func TestLunar(t *testing.T) {
	// 杨家大院建成日期：2005年3月初8 == 2005-4-16
	builtAt := calendar.ByLunar(2005, 3, 8, 0, 0, 0, false)
	y := builtAt.Solar.GetYear()
	m := builtAt.Solar.GetMonth()
	d := builtAt.Solar.GetDay()
	if !(y == 2005 && m == 4 && d == 16) {
		t.Fatal(`农历、阳历不相等。`)
	}
	{
		builtAt := calendar.ByLunar(2025, 6, 1, 0, 0, 0, true)
		y := builtAt.Solar.GetYear()
		m := builtAt.Solar.GetMonth()
		d := builtAt.Solar.GetDay()
		if !(y == 2025 && m == 7 && d == 25) {
			t.Fatal(`农历、阳历不相等。`)
		}
	}
}

func TestPrintLunar(t *testing.T) {
	cc := func(y, m, d int, leap bool) reminders.LunarDate {
		return reminders.NewLunarDate(y, m, d, 0, 0, 0, leap)
	}
	tests := []struct {
		l reminders.LunarDate
		s string
	}{
		{cc(2005, 3, 8, false), `二零零五年三月初八`},
		{cc(2005, 11, 12, false), `二零零五年冬月十二`},
		{cc(2005, 12, 20, false), `二零零五年腊月二十`},
		{cc(2005, 12, 23, false), `二零零五年腊月廿三`},
	}
	for _, test := range tests {
		if got := test.l.DateString(); got != test.s {
			t.Errorf(`%s != %s`, got, test.s)
		}
	}
}

func TestParseLunarDate(t *testing.T) {
	cc := func(y, m, d int, leap bool) reminders.LunarDate {
		return reminders.NewLunarDate(y, m, d, 0, 0, 0, leap)
	}
	tests := []struct {
		l reminders.LunarDate
		s string
	}{
		{cc(2005, 3, 8, false), `二零零五年三月初八`},
		{cc(2005, 11, 12, false), `二零零五年冬月十二`},
		{cc(2005, 12, 20, false), `二零零五年腊月二十`},
		{cc(2005, 12, 23, false), `二零零五年腊月廿三`},
		{cc(2025, 6, 1, true), `二零二五年闰六月初一`},
	}
	for _, test := range tests {
		want := test.l.DateString()
		lunar, err := reminders.ParseLunarDate(test.s)
		if err != nil {
			t.Errorf(`%s: %s`, test.s, want)
			continue
		}
		if got := lunar.DateString(); got != want {
			t.Errorf(`got:%s != want:%s`, got, want)
		}
	}
}

func TestLunarDateAddDays(t *testing.T) {
	l := reminders.NewLunarDate(2025, 6, 30, 0, 0, 0, false)
	a := l.AddDays(1)
	if a.DateString() != `二零二五年闰六月初一` {
		t.Fatalf(`农历不相等：%v`, a.DateString())
	}
}
