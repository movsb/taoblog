package calendar

import (
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

// 计算日期天数之差。
//
// 如果是未来时间，则是倒计时。
// 如果是过去时间，则是纪念日。
//
//  1. 日期时间会被归零到当天凌晨再计算。
//  2. exclusive（排除今天） 对未来时间无效。
//
// 示例：
//
//  1. now=2，t=1；结果：2 if !exclusive, 1 if exclusive
//  2. now=1，t=2；结果：1 跟 exclusive 无关。
func DaysPassed(now, t time.Time, exclusive bool) int {
	now = truncate(now)
	t = truncate(t)
	n := int(now.Sub(t).Hours() / 24)
	if n < 0 {
		return n
	}
	return utils.IIF(exclusive, n, n+1)
}

func truncate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func isAllDay(st, et time.Time) bool {
	startHasTime := st.Hour() != 0 || st.Minute() != 0 || st.Second() != 0
	endHasTime := et.Hour() != 0 || et.Minute() != 0 || et.Second() != 0
	return !(startHasTime || endHasTime)
}

// 每日任务。
// 把 start,end 时间扩展到 now 的一天范围内。
// end 可以跨天。
func Daily(now time.Time, start, end time.Time) (time.Time, time.Time) {
	s := time.Date(
		now.Year(), now.Month(), now.Day(),
		start.Hour(),
		start.Minute(),
		start.Second(),
		0, time.Local,
	)
	e := s.Add(end.Sub(start))
	return s, e
}

// 扩展为前 N 天。
// 如果为全天事件，只返回一组数据。
// 如果全天事件大于 1 天，参数无效，不处理。
func FirstDays(start, end time.Time, n int) [][2]time.Time {
	if isAllDay(start, end) {
		if end.Sub(start) > time.Hour*24 {
			return [][2]time.Time{
				{start, end},
			}
		}
		return [][2]time.Time{
			{start, start.AddDate(0, 0, n)},
		}
	}

	pairs := [][2]time.Time{}
	for i := range n {
		s := start.AddDate(0, 0, i)
		e := end.AddDate(0, 0, i)
		pairs = append(pairs, [2]time.Time{s, e})
	}
	return pairs
}

// 扩展为前 N 周。
// TODO: 如果为全天事件，只返回一组数据。
func FirstWeeks(start, end time.Time, n int) [][2]time.Time {
	pairs := [][2]time.Time{}
	for i := range n {
		s := start.AddDate(0, 0, 7*i)
		e := end.AddDate(0, 0, 7*i)
		pairs = append(pairs, [2]time.Time{s, e})
	}
	return pairs
}

// 添加自然月。
//
// 注意 AddDate：
//
// 2014-10-31 +1 个月，期待：2014-11-30，但实际会是 2014-12-01 号。
// 2014-12-31 +2 个月，期待：2015-02-28，但实际会是 2015-03-03 号。
//
// 实际结果均与目前的设计有违，手动往前调整到上个月最后一天。
//
// 注意，12月到1月会 round
func AddMonths(t time.Time, n int) time.Time {
	d2 := t.AddDate(0, n, 0)

	expect := int(t.Month()) + n
	if expect > 12 {
		expect -= 12
	}
	for expect != int(d2.Month()) {
		d2 = d2.AddDate(0, 0, -1)
	}
	return d2
}

func AddYears(t time.Time, n int) time.Time {
	d1 := t
	d2 := d1.AddDate(n, 0, 0)

	// 同上面月份的注意事项
	expect := int(d1.Month())
	for expect != int(d2.Month()) {
		d2 = d2.AddDate(0, 0, -1)
	}

	return d2
}
