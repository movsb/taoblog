package reminders

import (
	"fmt"
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
}

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

// userID 可能来自分享。
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
			st := r.Dates.Start.AddDate(0, 0, utils.IIF(r.Exclusive, day, day-1))
			et := r.Dates.End.AddDate(0, 0, utils.IIF(r.Exclusive, day, day-1))
			e := calendar.Event{
				Message: fmt.Sprintf(`%s 已经 %d 天了`, r.Title, day),

				Start: st,
				End:   et,

				UserID: userID,
				PostID: postID,

				Tags: map[string]any{
					`uuid`:   r.uuid,
					`repeat`: fmt.Sprintf(`day@%d`, day),
				},
			}
			s.cal.AddEvent(&e)
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
			st := r.Dates.Start.AddDate(0, 0, 7*week)
			et := r.Dates.End.AddDate(0, 0, 7*week)
			e := calendar.Event{
				Message: fmt.Sprintf(`%s 已经 %d 周了`, r.Title, week),

				Start: st,
				End:   et,

				UserID: userID,
				PostID: postID,

				Tags: map[string]any{
					`uuid`:   r.uuid,
					`repeat`: fmt.Sprintf(`week@%d`, week),
				},
			}
			s.cal.AddEvent(&e)
		}
	}

	for _, month := range r.Remind.Months {
		st := calendar.AddMonths(r.Dates.Start.Time, month)
		et := calendar.AddMonths(r.Dates.End.Time, month)

		e := calendar.Event{
			Message: fmt.Sprintf(`%s 已经 %d 个月了`, r.Title, month),

			Start: st,
			End:   et,

			UserID: userID,
			PostID: postID,

			Tags: map[string]any{
				`uuid`:   r.uuid,
				`repeat`: fmt.Sprintf(`month@%d`, month),
			},
		}
		s.cal.AddEvent(&e)
	}

	for _, year := range r.Remind.Years {
		if year < 1 {
			return fmt.Errorf(`提醒年份不能小于 1 年`)
		}

		st := calendar.AddYears(r.Dates.Start.Time, year)
		et := calendar.AddYears(r.Dates.End.Time, year)

		e := calendar.Event{
			Message: fmt.Sprintf(`%s 已经 %d 年了`, r.Title, year),

			Start: st,
			End:   et,

			UserID: userID,
			PostID: postID,

			Tags: map[string]any{
				`uuid`:   r.uuid,
				`repeat`: fmt.Sprintf(`year@%d`, year),
			},
		}
		s.cal.AddEvent(&e)
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
