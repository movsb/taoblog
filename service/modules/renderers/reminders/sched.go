package reminders

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/go-co-op/gocron/v2"
	"github.com/jonboulle/clockwork"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
)

type Job struct {
	underlying gocron.Job
	startAt    time.Time
	message    string
}

type Scheduler struct {
	backend gocron.Scheduler
	clock   clockwork.Clock

	lock sync.Mutex
	jobs map[int][]Job
}

type SchedulerOption func(s *Scheduler)

func WithFakeClock(clock clockwork.Clock) SchedulerOption {
	return func(s *Scheduler) {
		s.clock = clock
	}
}

func NewScheduler(options ...SchedulerOption) *Scheduler {
	sched := &Scheduler{
		jobs: make(map[int][]Job),
	}

	for _, opt := range options {
		opt(sched)
	}

	if sched.clock == nil {
		sched.clock = clockwork.NewRealClock()
	}

	backendOptions := []gocron.SchedulerOption{}
	if sched.clock != nil {
		backendOptions = append(backendOptions, gocron.WithClock(sched.clock))
	}

	sched.backend = utils.Must1(gocron.NewScheduler(backendOptions...))
	sched.backend.Start()

	return sched
}

func (s *Scheduler) AddReminder(postID int, r *Reminder, remind func(now time.Time, message string)) error {
	now := s.clock.Now()

	createJob := func(t time.Time, message string) error {
		log.Println(`创建任务：`, message)
		j, err := s.backend.NewJob(
			gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(t)),
			gocron.NewTask(remind, t, message),
			gocron.WithTags(
				fmt.Sprintf(`post_id:%d`, postID),
			),
		)
		if err != nil {
			log.Println(r, err)
			return err
		}

		s.lock.Lock()
		defer s.lock.Unlock()
		s.jobs[postID] = append(s.jobs[postID], Job{
			underlying: j,
			startAt:    t,
			message:    message,
		})

		return nil
	}

	for _, day := range r.Remind.Days {
		if day == 1 {
			return fmt.Errorf(`提醒天数不能为第 1 天`)
		}
		t := time.Time(r.Dates.Start).AddDate(0, 0, utils.IIF(r.Exclusive, day, day-1))
		if t.Before(now) {
			continue
		}

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

		if d2.Before(now) {
			continue
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

		if d2.Before(now) {
			continue
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
	s.backend.RemoveByTags(fmt.Sprintf(`post_id:%d`, id))
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.jobs, id)
}

func (s *Scheduler) ForEachPost(fn func(id int, jobs []Job)) {
	s.lock.Lock()
	defer s.lock.Unlock()

	for id, jobs := range s.jobs {
		fn(id, jobs)
	}
}

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

func (s *CalenderService) marshal(w io.Writer) error {
	cal := ics.NewCalendarFor(version.Name)
	cal.SetMethod(ics.MethodPublish)
	// TODO 写死了
	cal.SetTimezoneId(`Asia/Shanghai`)
	cal.SetXWRCalName(s.name)

	s.sched.ForEachPost(func(id int, jobs []Job) {
		for _, job := range jobs {
			eventID := fmt.Sprintf(`post_id:%d,job_id:%d`, id, job.startAt.Unix())
			e := cal.AddEvent(eventID)
			e.SetSummary(job.message)
			e.SetDtStampTime(job.startAt)
			e.SetStartAt(job.startAt)
			e.SetEndAt(job.startAt.AddDate(0, 0, 1).Add(-1))
			_ = e
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
	if err := s.marshal(w); err != nil {
		log.Println(err)
	}
}
