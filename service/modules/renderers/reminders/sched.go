package reminders

import (
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
)

type Job struct {
	startAt UserDate
	endAt   UserDate

	message     string
	messageFunc func() string

	// 以下选一个有效。
	isDaily    bool
	firstDays  any
	firstWeeks any
}

func (j Job) isAllDay() bool {
	st, et := j.startAt, j.endAt
	startHasTime := st.Hour() != 0 || st.Minute() != 0 || st.Second() != 0
	endHasTime := et.Hour() != 0 || et.Minute() != 0 || et.Second() != 0
	return !(startHasTime || endHasTime)
}

func (j Job) Message() string {
	if j.messageFunc != nil {
		return j.messageFunc()
	}
	return j.message
}

type Scheduler struct {
	lock   sync.Mutex
	jobs   map[int][]Job
	firsts map[int][]Job

	now func() time.Time
}

type SchedulerOption func(s *Scheduler)

func WithNowFunc(now func() time.Time) func(s *Scheduler) {
	return func(s *Scheduler) {
		s.now = now
	}
}

func NewScheduler(options ...SchedulerOption) *Scheduler {
	sched := &Scheduler{
		jobs:   make(map[int][]Job),
		firsts: make(map[int][]Job),
	}

	for _, opt := range options {
		opt(sched)
	}

	if sched.now == nil {
		sched.now = time.Now
	}

	return sched
}

func (s *Scheduler) withLock(fn func()) {
	s.lock.Lock()
	defer s.lock.Unlock()
	fn()
}

func (s *Scheduler) AddReminder(postID int, r *Reminder) error {
	omitToday := false

	if r.Remind.Daily {
		s.lock.Lock()
		s.firsts[postID] = append(s.firsts[postID], Job{
			startAt: r.Dates.Start,
			endAt:   r.Dates.End,
			messageFunc: func() string {
				days := daysPassed(s.now(), r.Dates.Start.Time, r.Exclusive)
				return fmt.Sprintf(`%s 已经 %d 天了`, r.Title, days)
			},
			isDaily: true,
		})
		s.lock.Unlock()
		// omitToday = true
	}

	createJob := func(t time.Time, message string) error {
		s.lock.Lock()
		defer s.lock.Unlock()
		s.jobs[postID] = append(s.jobs[postID], Job{
			startAt: UserDate{Time: t},
			message: message,
		})
		return nil
	}

	for _, day := range r.Remind.Days {
		if day == 1 {
			return fmt.Errorf(`提醒天数不能为第 1 天`)
		}

		if day < 0 {
			s.withLock(func() {
				s.firsts[postID] = append(s.firsts[postID], Job{
					startAt:   r.Dates.Start,
					endAt:     r.Dates.End,
					message:   r.Title,
					firstDays: -day,
				})
			})
			omitToday = true
			continue
		}

		t := r.Dates.Start.AddDate(0, 0, utils.IIF(r.Exclusive, day, day-1))

		if err := createJob(t, fmt.Sprintf(`%s 已经 %d 天了`, r.Title, day)); err != nil {
			return err
		}
	}

	for _, week := range r.Remind.Weeks {
		if week < 0 {
			s.withLock(func() {
				s.firsts[postID] = append(s.firsts[postID], Job{
					startAt:    r.Dates.Start,
					endAt:      r.Dates.End,
					message:    r.Title,
					firstWeeks: -week,
				})
			})
			omitToday = true
			continue
		}

		t := r.Dates.Start.AddDate(0, 0, week*7)
		if err := createJob(t, fmt.Sprintf(`%s 已经 %d 周了`, r.Title, week)); err != nil {
			return err
		}
	}

	for _, month := range r.Remind.Months {
		if month < 1 {
			return fmt.Errorf(`提醒月份不能小于 1 个月`)
		}

		d1 := r.Dates.Start
		d2 := d1.AddDate(0, month, 0)

		// 注意 AddDate：
		//
		// 2014-10-31 +1 个月，期待：2014-11-30，但实际会是 2014-12-01 号。
		// 2014-12-31 +2 个月，期待：2015-02-28，但实际会是 2015-03-03 号。
		//
		// 实际结果均与目前的设计有违，手动往前调整到上个月最后一天。
		//
		// 注意，12月到1月会 round
		expect := int(d1.Month()) + month
		if expect > 12 {
			expect -= 12
		}
		for expect != int(d2.Month()) {
			d2 = d2.AddDate(0, 0, -1)
		}

		if err := createJob(d2, fmt.Sprintf(`%s 已经 %d 个月了`, r.Title, month)); err != nil {
			return err
		}
	}

	for _, year := range r.Remind.Years {
		if year < 1 {
			return fmt.Errorf(`提醒年份不能小于 1 年`)
		}

		d1 := r.Dates.Start
		d2 := d1.AddDate(year, 0, 0)

		// 同上面月份的注意事项
		expect := int(d1.Month())
		for expect != int(d2.Month()) {
			d2 = d2.AddDate(0, 0, -1)
		}

		if err := createJob(d2, fmt.Sprintf(`%s 已经 %d 年了`, r.Title, year)); err != nil {
			return err
		}
	}

	if !omitToday {
		if err := createJob(r.Dates.Start.Time, r.Title); err != nil {
			return err
		}
	}

	log.Println(`提醒：处理完成：`, r.Title)

	return nil
}

// 根据文章编号删除提醒。
func (s *Scheduler) DeleteRemindersByPostID(id int) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.jobs, id)
	delete(s.firsts, id)
}

func (s *Scheduler) ForEachPost(fn func(id int, jobs []Job, firsts []Job)) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ids := make(map[int]struct{}, len(s.jobs)+len(s.firsts))
	for id := range s.jobs {
		ids[id] = struct{}{}
	}
	for id := range s.firsts {
		ids[id] = struct{}{}
	}

	for id := range ids {
		jobs := s.jobs[id]
		firsts := s.firsts[id]
		fn(id, jobs, firsts)
	}
}

// 用于把提醒事件导出为日历格式。
//
// https://icalendar.org/validator.html
type CalenderService struct {
	name  string
	sched *Scheduler
	*http.ServeMux
}

func NewCalendarService(name string, sched *Scheduler) *CalenderService {
	s := &CalenderService{
		name:     name,
		sched:    sched,
		ServeMux: http.NewServeMux(),
	}
	s.Handle(`/all.ics`, s.addHeaders(s.all))
	return s
}

// TODO: 对于每天任务，需要强制刷新修改时间以使得每日更新。
func (s *CalenderService) Marshal(now time.Time, w io.Writer) error {
	cal := ics.NewCalendarFor(version.Name)
	cal.SetMethod(ics.MethodPublish)
	cal.SetLastModified(now)
	// TODO 写死了
	cal.SetTimezoneId(`Asia/Shanghai`)
	cal.SetXWRCalName(s.name)

	s.sched.ForEachPost(func(id int, jobs []Job, firsts []Job) {
		for _, job := range jobs {
			eventID := fmt.Sprintf(
				`post_id:%d,job_id:%d,title:%x`,
				id, job.startAt.Unix(), crc32.ChecksumIEEE([]byte(job.Message())),
			)
			e := cal.AddEvent(eventID)
			e.SetSummary(job.Message())
			e.SetDtStampTime(job.startAt.Time)

			// 默认为全天事件
			e.SetAllDayStartAt(job.startAt.Time)
			// 不包含结束日。
			e.SetAllDayEndAt(job.startAt.AddDate(0, 0, 1))
		}

		for _, job := range firsts {
			switch {
			case job.isDaily:
				eventID := fmt.Sprintf(`post_id:%d,job_id:%d,daily:true`, id, job.startAt.Unix())
				e := cal.AddEvent(eventID)
				e.SetSummary(job.Message())
				e.SetDtStampTime(job.startAt.Time)

				if job.isAllDay() {
					e.SetAllDayStartAt(now)
					e.SetAllDayEndAt(now.AddDate(0, 0, 1))
				} else {
					t := time.Date(
						now.Year(), now.Month(), now.Day(),
						job.startAt.Hour(),
						job.startAt.Minute(),
						job.startAt.Second(),
						0, time.Local,
					)
					e.SetStartAt(t)
					e.SetEndAt(t.AddDate(0, 0, 1))
				}
			case job.firstDays != nil:
				days := job.firstDays.(int)

				if job.isAllDay() {
					eventID := fmt.Sprintf(`post_id:%d,job_id:%d,first_days:%d`, id, job.startAt.Unix(), days)
					e := cal.AddEvent(eventID)
					e.SetSummary(job.Message())
					e.SetDtStampTime(job.startAt.Time)

					e.SetAllDayStartAt(job.startAt.Time)
					e.SetAllDayEndAt(job.startAt.AddDate(0, 0, days))
				} else {
					for i := 1; i <= days; i++ {
						eventID := fmt.Sprintf(`post_id:%d,job_id:%d,first_days:%d:day:%d`, id, job.startAt.Unix(), days, i)
						e := cal.AddEvent(eventID)
						e.SetSummary(job.Message())
						e.SetDtStampTime(job.startAt.Time)
						e.SetStartAt(job.startAt.Time)
						e.SetEndAt(job.endAt.AddDate(0, 0, i))
					}
				}
			case job.firstWeeks != nil:
				weeks := job.firstWeeks.(int)

				if job.isAllDay() {
					eventID := fmt.Sprintf(`post_id:%d,job_id:%d,first_weeks:%d`, id, job.startAt.Unix(), weeks)
					e := cal.AddEvent(eventID)
					e.SetSummary(job.Message())
					e.SetDtStampTime(job.startAt.Time)
					e.SetAllDayStartAt(job.startAt.Time)
					e.SetAllDayEndAt(job.startAt.AddDate(0, 0, weeks))
				} else {
					for i := 1; i <= weeks; i++ {
						eventID := fmt.Sprintf(`post_id:%d,job_id:%d,first_weeks:%d,week:%d`, id, job.startAt.Unix(), weeks, i)
						e := cal.AddEvent(eventID)
						e.SetSummary(job.Message())
						e.SetDtStampTime(job.startAt.Time)
						e.SetStartAt(job.startAt.Time.AddDate(0, 0, 7*(i-1)))
						e.SetEndAt(job.endAt.AddDate(0, 0, 7*(i-1)))
					}
				}
			}
		}
	})

	return cal.SerializeTo(w, ics.WithNewLine("\r\n"))
}

func (s *CalenderService) addHeaders(h http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(`Content-Type`, `text/calendar; charset=utf-8`)
		h.ServeHTTP(w, r)
	})
}

func (s *CalenderService) all(w http.ResponseWriter, r *http.Request) {
	if err := s.Marshal(time.Now(), w); err != nil {
		log.Println(err)
	}
}
