package reminders

import (
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/calendar"
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
	r      *Reminder
	userID int
	postID int
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

func (s *Scheduler) UpdateDaily(postID int) {
	now := s.now()

	s.cal.Remove(func(e *calendar.Event) bool {
		if e.PostID == postID {
			_, isDaily := e.Tags[`daily`]
			return isDaily
		}
		return false
	})

	s.lock.Lock()
	defer s.lock.Unlock()

	for _, daily := range s.daily {
		if daily.postID != postID {
			continue
		}

		r := daily.r

		days := calendar.DaysPassed(now, r.Dates.Start.Time, r.Exclusive)
		st, et := calendar.Daily(now, r.Dates.Start.Time, r.Dates.End.Time)
		e := calendar.Event{
			Message: fmt.Sprintf(`%s 已经 %d 天了`, r.Title, days),

			Start: st,
			End:   et,

			UserID: daily.userID,
			PostID: daily.postID,

			Tags: map[string]any{
				`daily`: true,
				`uuid`:  r.uuid,
			},
		}
		s.cal.AddEvent(&e)
	}
}

func (s *Scheduler) UpdateEvery(postID int) {
	s.cal.Remove(func(e *calendar.Event) bool {
		if e.PostID == postID {
			_, isEvery := e.Tags[`every`]
			return isEvery
		}
		return false
	})

	s.lock.Lock()
	defer s.lock.Unlock()

	for _, every := range s.every {
		if every.postID != postID {
			continue
		}

		for _, d := range every.r.Remind.Every {
			dd, err := calendar.ParseDuration(d)
			if err != nil {
				log.Println(err)
				continue
			}
			s.addEvery(every.r, every.postID, every.userID, dd, nil)
		}
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
	s.cal.AddEvent(&e)
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
	s.cal.AddEvent(&e)
}

func (s *Scheduler) addMonth(r *Reminder, pid, uid int, month int, tags map[string]any) {
	st := calendar.AddMonths(r.Dates.Start.Time, month)
	et := calendar.AddMonths(r.Dates.End.Time, month)

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
	s.cal.AddEvent(&e)
}

func (s *Scheduler) addYear(r *Reminder, pid, uid int, year int, tags map[string]any) {
	st := calendar.AddYears(r.Dates.Start.Time, year)
	et := calendar.AddYears(r.Dates.End.Time, year)

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
	s.cal.AddEvent(&e)
}

func (s *Scheduler) addEvery(r *Reminder, pid, uid int, d calendar.Duration, shouldCreateToday *bool) {
	now := s.now()
	st := r.Dates.Start.Time
	tags := map[string]any{
		`every`: true,
	}
	switch d.Unit {
	case calendar.UnitDay:
		n := d.N
		for {
			st = st.AddDate(0, 0, utils.IIF(r.Exclusive, n, n-1))
			if st.After(now) {
				s.addDay(r, pid, uid, n, tags)
				if n == 1 && shouldCreateToday != nil {
					*shouldCreateToday = false
				}
				break
			}
			n *= 2
		}
	case calendar.UnitWeek:
		n := d.N
		for {
			st = st.AddDate(0, 0, 7*n)
			if st.After(now) {
				s.addWeek(r, pid, uid, n, tags)
				break
			}
			n *= 2
		}
	case calendar.UnitMonth:
		n := d.N
		for {
			st = calendar.AddMonths(st, n)
			if st.After(now) {
				s.addMonth(r, pid, uid, n, tags)
				break
			}
			n *= 2
		}
	case calendar.UnitYear:
		n := d.N
		for {
			st = calendar.AddYears(st, n)
			if st.After(now) {
				s.addYear(r, pid, uid, n, tags)
				break
			}
			n *= 2
		}
	}
}

// userID 可能来自分享。
// TODO 忽略过去的事件。
func (s *Scheduler) AddReminder(postID int, userID int, r *Reminder) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if r.Remind.Daily {
		s.daily = append(s.daily, _Daily{
			postID: postID,
			userID: userID,
			r:      r,
		})
	}
	if len(r.Remind.Every) > 0 {
		s.every = append(s.every, _Daily{
			postID: postID,
			userID: userID,
			r:      r,
		})
	}

	shouldCreateToday := true

	for _, day := range r.Remind.Days {
		if day == 1 {
			return fmt.Errorf(`提醒天数不能为第 1 天`)
		}

		if day < 0 {
			days := calendar.FirstDays(r.Dates.Start.Time, r.Dates.End.Time, -day)
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
				s.cal.AddEvent(&e)
			}
			shouldCreateToday = false
		} else {
			s.addDay(r, postID, userID, day, nil)
		}
	}

	for _, week := range r.Remind.Weeks {
		if week < 0 {
			weeks := calendar.FirstWeeks(r.Dates.Start.Time, r.Dates.End.Time, -week)
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
				s.cal.AddEvent(&e)
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

	for _, every := range r.Remind.Every {
		d, err := calendar.ParseDuration(every)
		if err != nil {
			return err
		}
		s.addEvery(r, postID, userID, d, &shouldCreateToday)
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
		s.cal.AddEvent(&e)
	}

	return nil
}

// 根据文章编号删除提醒。
func (s *Scheduler) DeleteRemindersByPostID(id int) {
	s.cal.Remove(func(e *calendar.Event) bool {
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
	s.cal.Each(func(e *calendar.Event) {
		ids[e.PostID] = struct{}{}
	})
	for pid := range ids {
		fn(pid)
	}
}
