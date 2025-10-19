package reminders

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/globals"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/service/modules/calendar"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/yuin/goldmark/parser"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed reminder.html style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `reminders`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

type RemindersOption func(*Reminders)

func WithOutputReminders(out *[]*Reminder) RemindersOption {
	return func(r *Reminders) {
		r.out = out
	}
}

func WithTimezone(location *time.Location) RemindersOption {
	return func(r *Reminders) {
		r.timezone = location
	}
}

type Reminders struct {
	sched *Scheduler

	out *[]*Reminder

	timezone *time.Location
}

func New(options ...RemindersOption) *Reminders {
	f := &Reminders{
		sched: NewScheduler(calendar.NewCalendarService(time.Now), time.Now),
		// TODO 传文章的时区进来。
		timezone: globals.SystemTimezone(),
	}

	for _, opt := range options {
		opt(f)
	}

	return f
}

var t = sync.OnceValue(func() *utils.TemplateLoader {
	return utils.NewTemplateLoader(utils.IIF(version.DevMode(), _root, fs.FS(_embed)), nil, dynamic.Reload)
})

func (r *Reminders) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) (outErr error) {
	defer utils.CatchAsError(&outErr)

	rm, err := ParseReminder(source, r.timezone)
	if err != nil {
		return fmt.Errorf(`解析提醒失败：%w`, err)
	}

	// 1：随便写的，因为 sched 是测试用的，没共享。
	if err := r.sched.AddReminder(1, 1, rm); err != nil {
		return fmt.Errorf(`添加提醒失败：%w`, err)
	}

	// TODO 在 Transform 的时候实现，以实现不渲染获取到数据。
	if r.out != nil {
		*r.out = append(*r.out, rm)
	}

	return t().GetNamed(`reminder.html`).Execute(w, rm)
}
