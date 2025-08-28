package solar

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
