package reminders

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/calendar"
	"github.com/movsb/taoblog/service/modules/calendar/solar"
)

type Scheduler struct {
	lock sync.Mutex
	now  func() time.Time
	cal  *calendar.CalenderService

	daily []_Daily
	every []_Daily
}

// TODO 重命名，同时适用于 every
type _Daily struct {
	reminder *Reminder
	userID   int
	postID   int
}

func NewScheduler(cal *calendar.CalenderService, now func() time.Time) *Scheduler {
	sched := &Scheduler{
		now: now,
		cal: cal,
	}

	if sched.now == nil {
		sched.now = time.Now
	}

	return sched
}

func (s *Scheduler) UpdateDaily() {
	s.lock.Lock()
	defer s.lock.Unlock()

	now := s.now()

	for _, daily := range s.daily {
		s.updateDaily(now, daily)
	}
}

var calKind = calendar.RegisterKind(func(e *calendar.Event) string {
	uuid := e.Tags[`uuid`].(string)
	repeat, _ := e.Tags[`repeat`].(string)
	return uuid + repeat
})

func (s *Scheduler) updateDaily(now time.Time, d _Daily) {
	r := d.reminder
	s.cal.Remove(calKind, func(e *calendar.Event) bool {
		uuid, _ := e.Tags[`uuid`]
		_, isDaily := e.Tags[`daily`]
		return uuid == r.uuid && isDaily
	})

	days := solar.DaysPassed(now, r.Dates.Start.Time, r.Exclusive)
	st, et := solar.Daily(now, r.Dates.Start.Time, r.Dates.End.Time)
	e := calendar.Event{
		Message: fmt.Sprintf(`%s 已经 %d 天了`, r.Title, days),

		Start: st,
		End:   et,

		UserID: d.userID,
		PostID: d.postID,

		Tags: map[string]any{
			`daily`: true,
			`uuid`:  r.uuid,
		},
	}
	s.cal.AddEvent(calKind, &e)
}

func (s *Scheduler) UpdateEvery() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, every := range s.every {
		s.updateEvery(every, nil)
	}
}

func (s *Scheduler) updateEvery(d _Daily, shouldCreateToday *bool) {
	r := d.reminder
	s.cal.Remove(calKind, func(e *calendar.Event) bool {
		uuid, _ := e.Tags[`uuid`]
		_, isEvery := e.Tags[`every`]
		return uuid == r.uuid && isEvery
	})

	for _, duration := range r.Remind.Every {
		dd, err := calendar.ParseDuration(duration)
		if err != nil {
			log.Println(err)
			continue
		}
		s.addEvery(r, d.postID, d.userID, dd, shouldCreateToday)
	}
}

func setTags(e *calendar.Event, tags map[string]any) {
	if e.Tags == nil {
		e.Tags = map[string]any{}
	}
	for k, v := range tags {
		e.Tags[k] = v
	}
}

func (s *Scheduler) addDay(r *Reminder, pid, uid int, day int, tags map[string]any) {
	st := r.Dates.Start.AddDate(0, 0, utils.IIF(r.Exclusive, day, day-1))
	et := r.Dates.End.AddDate(0, 0, utils.IIF(r.Exclusive, day, day-1))
	e := calendar.Event{
		Message: fmt.Sprintf(`%s 已经 %d 天了`, r.Title, day),

		Start: st,
		End:   et,

		UserID: uid,
		PostID: pid,

		Tags: map[string]any{
			`uuid`:   r.uuid,
			`repeat`: fmt.Sprintf(`day@%d`, day),
		},
	}
	setTags(&e, tags)
	s.cal.AddEvent(calKind, &e)
}

func (s *Scheduler) addWeek(r *Reminder, pid, uid int, week int, tags map[string]any) {
	st := r.Dates.Start.AddDate(0, 0, 7*week)
	et := r.Dates.End.AddDate(0, 0, 7*week)
	e := calendar.Event{
		Message: fmt.Sprintf(`%s 已经 %d 周了`, r.Title, week),

		Start: st,
		End:   et,

		UserID: uid,
		PostID: pid,

		Tags: map[string]any{
			`uuid`:   r.uuid,
			`repeat`: fmt.Sprintf(`week@%d`, week),
		},
	}
	setTags(&e, tags)
	s.cal.AddEvent(calKind, &e)
}

func (s *Scheduler) addMonth(r *Reminder, pid, uid int, month int, tags map[string]any) {
	st := solar.AddMonths(r.Dates.Start.Time, month)
	et := solar.AddMonths(r.Dates.End.Time, month)

	e := calendar.Event{
		Message: fmt.Sprintf(`%s 已经 %d 个月了`, r.Title, month),

		Start: st,
		End:   et,

		UserID: uid,
		PostID: pid,

		Tags: map[string]any{
			`uuid`:   r.uuid,
			`repeat`: fmt.Sprintf(`month@%d`, month),
		},
	}
	setTags(&e, tags)
	s.cal.AddEvent(calKind, &e)
}

func (s *Scheduler) addYear(r *Reminder, pid, uid int, year int, tags map[string]any) {
	st := solar.AddYears(r.Dates.Start.Time, year)
	et := solar.AddYears(r.Dates.End.Time, year)

	e := calendar.Event{
		Message: fmt.Sprintf(`%s 已经 %d 年了`, r.Title, year),

		Start: st,
		End:   et,

		UserID: uid,
		PostID: pid,

		Tags: map[string]any{
			`uuid`:   r.uuid,
			`repeat`: fmt.Sprintf(`year@%d`, year),
		},
	}
	setTags(&e, tags)
	s.cal.AddEvent(calKind, &e)
}

func (s *Scheduler) addEvery(r *Reminder, pid, uid int, d calendar.Duration, shouldCreateToday *bool) {
	now := s.now()
	tags := map[string]any{
		`every`: true,
	}

	today := func(t time.Time) bool {
		return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
	}

	switch d.Unit {
	case calendar.UnitDay:
		for i := 1; ; i++ {
			n := d.N * i
			st := r.Dates.Start.AddDate(0, 0, utils.IIF(r.Exclusive, n, n-1))
			if !st.Before(now) || today(st) {
				s.addDay(r, pid, uid, n, tags)
				if n == 1 && shouldCreateToday != nil {
					*shouldCreateToday = false
				}
				break
			}
		}
	case calendar.UnitWeek:
		for i := 1; ; i++ {
			n := d.N * i
			st := r.Dates.Start.AddDate(0, 0, 7*n)
			if !st.Before(now) || today(st) {
				s.addWeek(r, pid, uid, n, tags)
				break
			}
		}
	case calendar.UnitMonth:
		for i := 1; ; i++ {
			n := d.N * i
			st := solar.AddMonths(r.Dates.Start.Time, n)
			if !st.Before(now) || today(st) {
				s.addMonth(r, pid, uid, n, tags)
				break
			}
		}
	case calendar.UnitYear:
		for i := 1; ; i++ {
			n := d.N * i
			st := solar.AddYears(r.Dates.Start.Time, n)
			if !st.Before(now) || today(st) {
				s.addYear(r, pid, uid, n, tags)
				break
			}
		}
	}
}

// userID 可能来自分享。
// TODO 忽略过去的事件。
func (s *Scheduler) AddReminder(postID int, userID int, r *Reminder) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	shouldCreateToday := true

	if r.Remind.Daily {
		d := _Daily{
			postID:   postID,
			userID:   userID,
			reminder: r,
		}
		s.daily = append(s.daily, d)
		s.updateDaily(s.now(), d)
		shouldCreateToday = false
	}
	if len(r.Remind.Every) > 0 {
		d := _Daily{
			postID:   postID,
			userID:   userID,
			reminder: r,
		}
		s.every = append(s.every, d)
		s.updateEvery(d, &shouldCreateToday)
	}

	for _, day := range r.Remind.Days {
		if day == 1 {
			return fmt.Errorf(`提醒天数不能为第 1 天`)
		}

		if day < 0 {
			days := solar.FirstDays(r.Dates.Start.Time, r.Dates.End.Time, -day)
			for i, pair := range days {
				e := calendar.Event{
					Message: fmt.Sprintf(`%s`, r.Title),

					Start: pair[0],
					End:   pair[1],

					UserID: userID,
					PostID: postID,

					Tags: map[string]any{
						`uuid`:   r.uuid,
						`repeat`: fmt.Sprintf(`day@%d`, i),
					},
				}
				s.cal.AddEvent(calKind, &e)
			}
			shouldCreateToday = false
		} else {
			s.addDay(r, postID, userID, day, nil)
		}
	}

	for _, week := range r.Remind.Weeks {
		if week < 0 {
			weeks := solar.FirstWeeks(r.Dates.Start.Time, r.Dates.End.Time, -week)
			for i, pair := range weeks {
				e := calendar.Event{
					Message: r.Title,

					Start: pair[0],
					End:   pair[1],

					UserID: userID,
					PostID: postID,

					Tags: map[string]any{
						`uuid`:   r.uuid,
						`repeat`: fmt.Sprintf(`week@%d`, i),
					},
				}
				s.cal.AddEvent(calKind, &e)
			}
			shouldCreateToday = false
		} else {
			s.addWeek(r, postID, userID, week, nil)
		}
	}

	for _, month := range r.Remind.Months {
		s.addMonth(r, postID, userID, month, nil)
	}

	for _, year := range r.Remind.Years {
		if year < 1 {
			return fmt.Errorf(`提醒年份不能小于 1 年`)
		}

		s.addYear(r, postID, userID, year, nil)
	}

	if shouldCreateToday {
		e := calendar.Event{
			Message: r.Title,

			Start: r.Dates.Start.Time,
			End:   r.Dates.End.Time,

			UserID: userID,
			PostID: postID,

			Tags: map[string]any{
				`uuid`: r.uuid,
			},
		}
		s.cal.AddEvent(calKind, &e)
	}

	return nil
}

// 根据文章编号删除提醒。
func (s *Scheduler) DeleteRemindersByPostID(id int) {
	s.cal.Remove(calKind, func(e *calendar.Event) bool {
		return e.PostID == id
	})
	s.lock.Lock()
	defer s.lock.Unlock()

	s.daily = slices.DeleteFunc(s.daily, func(d _Daily) bool {
		return d.postID == id
	})
	s.every = slices.DeleteFunc(s.every, func(d _Daily) bool {
		return d.postID == id
	})
}

func (s *Scheduler) ForEachPost(fn func(id int)) {
	ids := map[int]struct{}{}
	s.cal.Each(calKind, func(e *calendar.Event) {
		ids[e.PostID] = struct{}{}
	})
	for pid := range ids {
		fn(pid)
	}
}
