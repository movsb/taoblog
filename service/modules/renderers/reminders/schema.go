package reminders

import (
	"fmt"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/google/uuid"
	"github.com/movsb/taoblog/service/modules/calendar"
)

type UserDate struct {
	time.Time
}

var layouts = [...]string{
	`2006-01-02`,
	`2006-01-02 15:04`,
}

// TODO 需要修改成服务器时间。
var FixedZone = time.FixedZone(`fixed`, 8*60*60)

func NewUserDateFromString(s string) (UserDate, error) {
	outErr := fmt.Errorf(`无法解析日期：%v`, s)
	for _, layout := range layouts {
		// TODO 需要区分 Parse 与 ParseInLocation
		t, err := time.ParseInLocation(layout, s, FixedZone)
		if err != nil {
			outErr = err
			continue
		}
		return UserDate{Time: t}, nil
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

	// * 对于发生在过去的事件时间，是否排除当天。
	// * 如果正值事件当天，此参数也有效。
	// * 如果事件还未发生，此参数无效。
	Exclusive bool `yaml:"exclusive"`

	// 提醒？
	Remind ReminderRemind `yaml:"remind"`

	// 文件共享后日历也可以共享给不同的用户
	// 为了显示结果去重，用 uuid 标识给 event 分组。
	uuid string
}

type ReminderDates struct {
	// 开始时间：日期和时间 或 日期
	Start UserDate `yaml:"start"`
	// 结束时间：日期和时间 或 日期，可选。
	End UserDate `yaml:"end"`
}

// NOTE: 日、周、月、年均支持负数，分别表示“前多少日、周、月、年”。
type ReminderRemind struct {
	// 序数词，表示第几天。
	// 当天可包含在内，也可不包含在内；由 Exclusive 决定。
	Days []int `yaml:"days"`

	// 第多少周
	//
	// 处理成当天所在星期。
	//
	// 比如今天是星期二，则依次为：下周二，下下周二。
	//
	// 适用于课程表。
	Weeks []int `yaml:"weeks"`

	// 第几个月。
	//
	// 对于类似 1.31 号这样的日期，目前的处理逻辑是：下个月是 2.28/2.29 。
	// 这也是苹果日历目前的处理方式。
	Months []int `yaml:"months"`

	// 第几年。
	Years []int `yaml:"years"`

	// 每天提醒。用于计数，比如“分开了多少天了”这样的。
	//
	// 注意：
	//  - 此事件不创建提醒。
	//  - 此事件只会创建于今日。
	//
	// 即：仅用于日历展示。
	Daily bool `yaml:"daily"`

	// 每……
	// [1d,2w,3m,4y]
	// 每1天、每2周、每3个月、每4年
	// 只保留下一个事件。
	Every []string `yaml:"every"`

	// TODO not used
	First int `yaml:"first"`
}

func (r *Reminder) Days() int {
	return calendar.DaysPassed(time.Now(), r.Dates.Start.Time, r.Exclusive)
}

func (r *Reminder) Start() string {
	return r.Dates.Start.Format(`2006-01-02`)
}

func ParseReminder(y []byte) (*Reminder, error) {
	rm := Reminder{}
	if err := yaml.UnmarshalWithOptions(y, &rm, yaml.Strict()); err != nil {
		return nil, err
	}
	rm.uuid = uuid.NewString()
	// 如果不指定结束，默认为全天事件。
	if rm.Dates.End.IsZero() {
		rm.Dates.End.Time = rm.Dates.Start.AddDate(0, 0, 1)
	}
	return &rm, nil
}
