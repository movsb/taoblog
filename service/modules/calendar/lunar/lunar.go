package lunar

import (
	"fmt"
	"time"

	"github.com/Lofanmi/chinese-calendar-golang/calendar"
)

type SolarDate time.Time

// 农历日期，及相关计算。
type LunarDate struct {
	c *calendar.Calendar
}

func NewLunarDate(year, month, day int, hour, minute, second int, leapMonth bool) LunarDate {
	return LunarDate{
		c: calendar.ByLunar(int64(year), int64(month), int64(day), int64(hour), int64(minute), int64(second), leapMonth),
	}
}

func (d LunarDate) IsZero() bool {
	return d.c == nil || d.SolarTime().IsZero()
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
//
// 添加的是自然年，也是是 N 年前是/后的同一天（或最接近）。
// 由于存在闰月，所以会返回 1~2 个结果，前者是不闰的，后者是闰的（如果有闰月的话）。
func (d LunarDate) AddYears(n int) []LunarDate {
	var (
		l             = d.c.Lunar
		expectedYear  = int(l.GetYear()) + n
		expectedMonth = l.GetMonth()
		expectedDay   = l.GetDay()
	)

	// 计算给定的时间是否是符合上述期望的时间
	// 第二个返回值表示是否是闰月。
	expected := func(near time.Time) (bool, bool) {
		lu := LunarDate{c: calendar.ByTimestamp(near.Unix())}
		gotYear := lu.c.Lunar.GetYear()
		gotMonth := lu.c.Lunar.GetMonth()
		isLeapMonth := lu.c.Lunar.IsLeapMonth()
		gotDay := lu.c.Lunar.GetDay()

		if expectedYear == int(gotYear) && expectedMonth == gotMonth {
			if expectedDay == gotDay {
				return true, isLeapMonth
			}
		}
		return false, false
	}

	// 先根据阳历计算出一个模糊的 N 年后的日子。
	base := d.SolarTime().AddDate(n, 0, 0)

	// 农历平年比阳历短约 11 天（365 − 354 ≈ 11）。
	// 前后计算大概 100 天以推算到最接近的一天。

	var (
		notLeap time.Time
		leap    time.Time
	)

retry:
	for i := range 100 {
		near1 := base.AddDate(0, 0, i)
		if ok, l := expected(near1); ok {
			if l {
				leap = near1
			} else {
				notLeap = near1
			}
		}

		near2 := base.AddDate(0, 0, -i)
		if ok, l := expected(near2); ok {
			if l {
				leap = near2
			} else {
				notLeap = near2
			}
		}
	}

	// 没找到。
	if notLeap.IsZero() && leap.IsZero() {
		// 有可能是因为是月末，但是另一年没有该天。
		if d.AddDays(1).c.Lunar.GetDay() == 1 {
			expectedDay--
			goto retry
		} else {
			panic(fmt.Sprintf(`找不到对应的农历：%s + %d Year`, d.DateString(), n))
		}
	}

	if notLeap.IsZero() {
		panic(`没有找到对应的农历`)
	}

	ret := []LunarDate{{c: calendar.ByTimestamp(notLeap.Unix())}}
	if !leap.IsZero() {
		ret = append(ret, LunarDate{c: calendar.ByTimestamp(leap.Unix())})
	}

	return ret
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
