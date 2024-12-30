package reminders

import "time"

type UserDate time.Time

var layouts = [...]string{
	`2006-01-02`,
}

// TODO 需要修改成服务器时间。
var fixedZone = time.FixedZone(`fixed`, 8*60*60)

func NewUserDateFromString(s string) (UserDate, error) {
	var outErr error
	for _, layout := range layouts {
		// TODO 需要区分 Parse 与 ParseInLocation
		t, err := time.ParseInLocation(layout, s, fixedZone)
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

type ReminderRemind struct {
	// 序数词，表示第 ？天。
	Days []int `yaml:"days"`
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
