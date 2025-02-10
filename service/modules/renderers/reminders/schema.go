package reminders

import (
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

type UserDate time.Time

var layouts = [...]string{
	`2006-01-02`,
}

// TODO 需要修改成服务器时间。
var FixedZone = time.FixedZone(`fixed`, 8*60*60)

func NewUserDateFromString(s string) (UserDate, error) {
	var outErr error
	for _, layout := range layouts {
		// TODO 需要区分 Parse 与 ParseInLocation
		t, err := time.ParseInLocation(layout, s, FixedZone)
		if err != nil {
			outErr = err
			continue
		}
		return UserDate(t), nil
	}
	return UserDate{}, outErr
}

func (u *UserDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	d, err := NewUserDateFromString(s)
	if err != nil {
		return err
	}
	*u = d
	return nil
}

type Reminder struct {
	Title       string        `yaml:"title"`
	Description string        `yaml:"description"`
	Dates       ReminderDates `yaml:"dates"`

	Exclusive bool `yaml:"exclusive"` // 排除今天？

	// 提醒？
	Remind ReminderRemind `yaml:"remind"`
}

type ReminderDates struct {
	Start UserDate `yaml:"start"`
}

func DateStart(s string) ReminderDates {
	return ReminderDates{
		Start: utils.Must1(NewUserDateFromString(s)),
	}
}

type ReminderRemind struct {
	// 序数词，表示第几天。
	// 当天可包含在内，也可不包含在内；由 Exclusive 决定。
	Days []int `yaml:"days"`

	// 第几个月。
	//
	// 对于类似 1.31 号这样的日期，目前的处理逻辑是：下个月是 2.28/2.29 。
	// 这也是苹果日历目前的处理方式。
	Months []int `yaml:"months"`

	// 第几年。
	Years []int `yaml:"years"`
}

func (r *Reminder) Days() int {
	// NOTE: +1 表示包含当天。
	n := int(time.Since(time.Time(r.Dates.Start)).Hours() / 24)
	if r.Exclusive {
		return n
	}
	return n + 1
}

func (r *Reminder) Start() string {
	return time.Time(r.Dates.Start).Format(`2006-01-02`)
}
