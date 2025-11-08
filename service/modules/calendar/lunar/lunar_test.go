package lunar_test

import (
	"fmt"
	"testing"

	"github.com/Lofanmi/chinese-calendar-golang/calendar"
	"github.com/movsb/taoblog/service/modules/calendar/lunar"
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
	cc := func(y, m, d int, leap bool) lunar.LunarDate {
		return lunar.NewLunarDate(y, m, d, 0, 0, 0, leap)
	}
	tests := []struct {
		l lunar.LunarDate
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
	cc := func(y, m, d int, leap bool) lunar.LunarDate {
		return lunar.NewLunarDate(y, m, d, 0, 0, 0, leap)
	}
	tests := []struct {
		l lunar.LunarDate
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
		lunar, err := lunar.ParseLunarDate(test.s)
		if err != nil {
			t.Errorf(`%s: %s`, test.s, want)
			continue
		}
		if got := lunar.DateString(); got != want {
			t.Errorf(`got:%s != want:%s`, got, want)
		}
	}
}

// 在 GitHub Actions 上面偶尔可能跑不过，暂时不测试。
func TestLunarDateAddDays(t *testing.T) {
	l := lunar.NewLunarDate(2025, 6, 30, 0, 0, 0, false)
	// t.Log(`当前：`, l.SolarTime().Unix())
	a := l.AddDays(1)
	// t.Log(`加一：`, a.SolarTime().Unix())
	if a.DateString() != `二零二五年闰六月初一` {
		t.Fatalf(`农历不相等：%v`, a.DateString())
	}
}

func N(ll []lunar.LunarDate, n int) []lunar.LunarDate {
	if len(ll) != n {
		panic(`农历个数不一样。`)
	}
	return ll
}

func LE(l lunar.LunarDate, s string) {
	if l.DateString() != s {
		panic(fmt.Sprintf(`农历不相等：%s, %s`, l.DateString(), s))
	}
}

func TestLunarDateAddYears(t *testing.T) {
	l := lunar.NewLunarDate(1991, 2, 20, 0, 0, 0, false)

	ll := N(l.AddYears(1), 1)
	LE(ll[0], `一九九二年二月二十`)

	// 2004 年含闰月。
	ll = N(l.AddYears(13), 2)
	LE(ll[0], `二零零四年二月二十`)
	LE(ll[1], `二零零四年闰二月二十`)

	// 2091 年，100 岁生日。
	ll = N(l.AddYears(100), 1)
	LE(ll[0], `二零九一年二月二十`)
}

func TestLunarDateAddYears2(t *testing.T) {
	l := lunar.NewLunarDate(2005, 2, 30, 0, 0, 0, false)
	ll := N(l.AddYears(1), 1)
	LE(ll[0], `二零零六年二月廿九`)
}
