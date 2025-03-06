package reminders

import (
	"embed"
	"html/template"
	"io"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/theme/modules/sass"
	"github.com/yuin/goldmark/parser"
)

//go:generate sass --no-source-map style.scss style.css

//go:embed reminder.html style.css
var _root embed.FS

func init() {
	dynamic.RegisterInit(func() {
		dynamic.Dynamic[`reminders`] = dynamic.Content{
			Styles: []string{
				string(utils.Must1(_root.ReadFile(`style.css`))),
			},
		}
		sass.WatchDefaultAsync(string(dir.SourceAbsoluteDir()))
	})
}

type RemindersOption func(*Reminders)

func WithOutputReminders(out *[]*Reminder) RemindersOption {
	return func(r *Reminders) {
		r.out = out
	}
}

type Reminders struct {
	out *[]*Reminder
}

func New(options ...RemindersOption) *Reminders {
	f := &Reminders{}

	for _, opt := range options {
		opt(f)
	}

	return f
}

var tmpl = template.Must(template.New(`reminder`).Parse(string(utils.Must1(_root.ReadFile(`reminder.html`)))))

func (r *Reminders) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) error {
	rm, err := ParseReminder(source)
	if err != nil {
		return err
	}

	// TODO 在 Transform 的时候实现，以实现不渲染获取到数据。
	if r.out != nil {
		*r.out = append(*r.out, rm)
	}

	return tmpl.Execute(w, rm)
}
