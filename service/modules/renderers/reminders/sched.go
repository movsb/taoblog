package reminders

import (
	"fmt"
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
	startAt time.Time

	message     string
	messageFunc func() string

	isDaily   bool
	isFirst   bool
	firstDays int
}

func (j Job) Message() string {
	if j.messageFunc != nil {
		return j.messageFunc()
	}
	return j.message
}

type Scheduler struct {
	lock  sync.Mutex
	jobs  map[int][]Job
	daily map[int][]Job
}

type SchedulerOption func(s *Scheduler)

func NewScheduler(options ...SchedulerOption) *Scheduler {
	sched := &Scheduler{
		jobs:  make(map[int][]Job),
		daily: make(map[int][]Job),
	}

	for _, opt := range options {
		opt(sched)
	}

	return sched
}

func (s *Scheduler) AddReminder(postID int, r *Reminder) error {
	if r.Remind.Daily {
		s.lock.Lock()
		s.daily[postID] = append(s.daily[postID], Job{
			startAt: time.Time(r.Dates.Start),
			messageFunc: func() string {
				return fmt.Sprintf(`%s 已经 %d 天了`, r.Title, daysPassed(time.Time(r.Dates.Start), r.Exclusive))
			},
			isDaily: true,
		})
		s.lock.Unlock()
	}
	if r.Remind.First > 0 {
		s.lock.Lock()
		s.daily[postID] = append(s.daily[postID], Job{
			startAt:   time.Time(r.Dates.Start),
			message:   r.Title,
			isFirst:   true,
			firstDays: r.Remind.First,
		})
		s.lock.Unlock()
	}

	createJob := func(t time.Time, message string) error {
		// log.Println(`创建任务：`, message)

		s.lock.Lock()
		defer s.lock.Unlock()
		s.jobs[postID] = append(s.jobs[postID], Job{
			startAt: t,
			message: message,
		})

		return nil
	}

	// 始终创建当天提醒。
	// 但是 first 任务不需要
	if r.Remind.First <= 0 {
		if err := createJob(time.Time(r.Dates.Start), r.Title); err != nil {
			return err
		}
	}

	for _, day := range r.Remind.Days {
		if day == 1 {
			return fmt.Errorf(`提醒天数不能为第 1 天`)
		}

		t := time.Time(r.Dates.Start).AddDate(0, 0, utils.IIF(r.Exclusive, day, day-1))

		if err := createJob(t, fmt.Sprintf(`%s 已经 %d 天了`, r.Title, day)); err != nil {
			return err
		}
	}

	for _, month := range r.Remind.Months {
		if month < 1 {
			return fmt.Errorf(`提醒月份不能小于 1 个月`)
		}

		d1 := time.Time(r.Dates.Start)
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

		d1 := time.Time(r.Dates.Start)
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

	log.Println(`提醒：添加任务：`, r.Title)

	return nil
}

// 根据文章编号删除提醒。
func (s *Scheduler) DeleteRemindersByPostID(id int) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.jobs, id)
	delete(s.daily, id)
}

func (s *Scheduler) ForEachPost(fn func(id int, jobs []Job, daily []Job)) {
	s.lock.Lock()
	defer s.lock.Unlock()

	ids := make(map[int]struct{}, len(s.jobs)+len(s.daily))
	for id := range s.jobs {
		ids[id] = struct{}{}
	}
	for id := range s.daily {
		ids[id] = struct{}{}
	}

	for id := range ids {
		jobs := s.jobs[id]
		daily := s.daily[id]
		fn(id, jobs, daily)
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

	s.sched.ForEachPost(func(id int, jobs []Job, daily []Job) {
		for _, job := range jobs {
			eventID := fmt.Sprintf(`post_id:%d,job_id:%d`, id, job.startAt.Unix())
			e := cal.AddEvent(eventID)
			e.SetSummary(job.Message())
			e.SetDtStampTime(job.startAt)

			// 默认为全天事件
			e.SetAllDayStartAt(job.startAt)
			// 不包含结束日。
			e.SetAllDayEndAt(job.startAt.AddDate(0, 0, 1))
		}

		for _, job := range daily {
			if job.isDaily {
				eventID := fmt.Sprintf(`post_id:%d,job_id:%d,daily:true`, id, job.startAt.Unix())
				e := cal.AddEvent(eventID)
				e.SetSummary(job.Message())
				e.SetDtStampTime(job.startAt)

				// 默认为全天事件
				e.SetAllDayStartAt(now)
				// 不包含结束日。
				e.SetAllDayEndAt(now.AddDate(0, 0, 1))
			} else if job.isFirst {
				eventID := fmt.Sprintf(`post_id:%d,job_id:%d,first:%d`, id, job.startAt.Unix(), job.firstDays)
				e := cal.AddEvent(eventID)
				e.SetSummary(job.Message())
				e.SetDtStampTime(job.startAt)
				e.SetAllDayStartAt(job.startAt)
				e.SetAllDayEndAt(job.startAt.AddDate(0, 0, job.firstDays))
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
