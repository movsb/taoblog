package lunar

import (
	"fmt"
	"time"

	"github.com/Lofanmi/chinese-calendar-golang/calendar"
)

// 农历日期，及相关计算。
type LunarDate struct {
	c *calendar.Calendar
}

func NewLunarDate(year, month, day int, hour, minute, second int, leapMonth bool) LunarDate {
	return LunarDate{
		c: calendar.ByLunar(int64(year), int64(month), int64(day), int64(hour), int64(minute), int64(second), leapMonth),
	}
}

// 返回形如“二零零五年三月初八”的农历日期显示。
func (d LunarDate) DateString() string {
	l := d.c.Lunar
	return fmt.Sprintf(`%s年%s%s`, l.YearAlias(), l.MonthAlias(), l.DayAlias())
}

// 返回对应的阳历时间。
func (d LunarDate) SolarTime() time.Time {
	s := d.c.Solar
	t := time.Date(
		int(s.GetYear()), time.Month(s.GetMonth()), int(s.GetDay()),
		int(s.GetHour()), int(s.GetMinute()), int(s.GetSecond()), int(s.GetNanosecond()),
		time.Local,
	)
	return t
}

// 返回“添加 N 天”后的农历日期。
func (d LunarDate) AddDays(n int) LunarDate {
	// AddDate 会因为时区产生偏移，所以这里直接计算。
	t := d.SolarTime().Add(time.Hour * 24 * time.Duration(n))
	c := calendar.ByTimestamp(t.Unix())
	return LunarDate{c: c}
}

// 返回“添加 N 年”后的农历日期。
// 对于月份，目前只保证数字相等，不包含“闰”。
func (d LunarDate) AddYears(n int) LunarDate {
	var (
		l             = d.c.Lunar
		expectedYear  = int(l.GetYear()) + n
		expectedMonth = l.GetMonth()
		expectedDay   = l.GetDay()
	)

	calc := func(near time.Time) bool {
		lu := LunarDate{c: calendar.ByTimestamp(near.Unix())}
		gotYear := lu.c.Lunar.GetYear()
		gotMonth := lu.c.Lunar.GetMonth()
		isLeapMonth := lu.c.Lunar.IsLeapMonth()
		gotDay := lu.c.Lunar.GetDay()

		if expectedYear == int(gotYear) && expectedMonth == gotMonth {
			if expectedDay == gotDay && !isLeapMonth {
				return true
			}
		}
		return false
	}

	base := d.SolarTime().AddDate(n, 0, 0)

	// 前后计算 1000 天以推算到最接近的一天。
	for i := range 1000 {
		near1 := base.AddDate(0, 0, i)
		if calc(near1) {
			return LunarDate{c: calendar.ByTimestamp(near1.Unix())}
		}

		near2 := base.AddDate(0, 0, -i)
		if calc(near2) {
			return LunarDate{c: calendar.ByTimestamp(near2.Unix())}
		}
	}

	// 最后一天天数可能不一样，前进到下一个月，然后减一天。
	if expectedDay != 1 && d.AddDays(1).c.Lunar.GetDay() == 1 {
		return d.AddDays(1).AddYears(n).AddDays(-1)
	}

	panic(fmt.Sprintf(`找不到对应的农历：%s + %d Year`, d.DateString(), n))
}

// 解析农历日期。
//
// 格式要求：
//
//   - 长度固定；
//   - 全部用中文书写✍️；
//   - 月份用一个汉字表示（11月为冬月，12月为腊月）；
//   - 1-10号用“初X”表示，11-20用对应汉字表示，21-29用“廿X”表示，30号用“三十”表示。
//   - 月份支持“闰”。
//
// 示例：
//
//   - 二零零五年三月初八
//   - 二零零五年冬月十二
//   - 二零零五年腊月二十
//   - 二零零五年腊月廿三
//   - 二零零五年闰二月廿三
func ParseLunarDate(s string) (*LunarDate, error) {
	var err = fmt.Errorf(`无法解析农历日期：%v`, s)
	var (
		year, month, day int64
		leap             bool
	)
	var (
		chars = []rune(s)
		i     int
	)

	if n := len(chars); n < 9 || n > 10 {
		return nil, err
	}

	// 失败返回-1
	digit := func(ty rune, ch rune) int64 {
		switch ch {
		case '一':
			return 1
		case '二':
			return 2
		case '三':
			return 3
		case '四':
			return 4
		case '五':
			return 5
		case '六':
			return 6
		case '七':
			return 7
		case '八':
			return 8
		case '九':
			return 9
		}

		switch ty {
		case '年':
			if ch == '零' {
				return 0
			}
		case '月':
			switch ch {
			case '冬':
				return 11
			case '腊':
				return 12
			}
		}

		return -1
	}

	for range 4 {
		d := digit('年', chars[i])
		if d == -1 {
			return nil, err
		}
		year *= 10
		year += d
		i++
	}

	if year < 1900 || year > 2100 {
		return nil, err
	}

	if chars[i] != '年' {
		return nil, err
	}
	i++

	if chars[i] == '闰' {
		leap = true
		i++
	}

	{
		d := digit('月', chars[i])
		if d == -1 {
			return nil, err
		}
		month = d
		i++
	}

	if chars[i] != '月' {
		return nil, err
	}
	i++

	switch chars[i] {
	case '初':
		i++
		if chars[i] == '十' {
			day = 10
		} else {
			day = digit('?', chars[i])
		}
		i++
	case '十':
		i++
		day = digit('?', chars[i])
		if day != -1 {
			day = 10 + day
		}
		i++
	case '二':
		i++
		if chars[i] == '十' {
			day = 20
			i++
		} else {
			day = -1
		}
	case '廿':
		i++
		day = digit('?', chars[i])
		if day != -1 {
			day = 20 + day
		}
		i++
	case '三':
		i++
		if chars[i] == '十' {
			day = 30
			i++
		} else {
			day = -1
		}
	}
	if day == -1 {
		return nil, err
	}

	l := calendar.ByLunar(year, month, day, 0, 0, 0, leap)
	return &LunarDate{c: l}, nil
}
