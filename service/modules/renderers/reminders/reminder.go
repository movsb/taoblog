package reminders

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/theme/modules/sass"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	"gopkg.in/yaml.v2"
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

type Reminders struct {
	task *Task
	pid  int
}

func New(task *Task, pid int) *Reminders {
	f := &Reminders{
		task: task,
		pid:  pid,
	}

	return f
}

type UserDate time.Time

var layouts = [...]string{
	`2006-01-02`,
}

// TODO 需要修改成服务器时间。
var fixedZone = time.FixedZone(`fixed`, 8*60*60)

func (u *UserDate) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	var outErr error
	for _, layout := range layouts {
		// TODO 需要区分 Parse 与 ParseInLocation
		t, err := time.ParseInLocation(layout, s, fixedZone)
		if err != nil {
			outErr = err
			continue
		}
		*u = UserDate(t)
	}
	return outErr
}

type Reminder struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Dates       struct {
		Start UserDate `yaml:"start"`
	} `yaml:"dates"`
}

func (r *Reminder) Days() int {
	return int(time.Since(time.Time(r.Dates.Start)).Hours()/24) + 1
}

func (r *Reminder) Start() string {
	return time.Time(r.Dates.Start).Format(`2006-01-02`)
}

var _ interface {
	parser.ASTTransformer
	goldmark.Extender
	renderer.NodeRenderer
} = (*Reminders)(nil)

var tmpl = template.Must(template.New(`reminder`).Parse(string(utils.Must1(_root.ReadFile(`reminder.html`)))))

func (r *Reminders) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(r, 100)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(r, 100)))
}

type _ReminderRendererBlock struct {
	ast.BaseBlock
	ref *ast.FencedCodeBlock
}

var _reminderCodeBLockKind = ast.NewNodeKind(`reminder_code_block`)

func (b *_ReminderRendererBlock) Kind() ast.NodeKind {
	return _reminderCodeBLockKind
}
func (b *_ReminderRendererBlock) Dump(source []byte, level int) {
	b.ref.Dump(source, level)
}

func (r *Reminders) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(_reminderCodeBLockKind, r.renderCodeBlock)
}

// Transform implements parser.ASTTransformer.
func (r *Reminders) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	rCodeBlocks := []*ast.FencedCodeBlock{}
	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Kind() == ast.KindFencedCodeBlock {
			cb := n.(*ast.FencedCodeBlock)
			if cb.Info != nil {
				info := string(cb.Info.Segment.Value(reader.Source()))
				if info == `reminder` {
					rCodeBlocks = append(rCodeBlocks, cb)
				}
			}
		}
		return ast.WalkContinue, nil
	})
	for _, cb := range rCodeBlocks {
		cb.Parent().ReplaceChild(cb.Parent(), cb, &_ReminderRendererBlock{
			ref: cb,
		})
	}
}

func (r *Reminders) renderCodeBlock(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n = n.(*_ReminderRendererBlock).ref
	b := bytes.NewBuffer(nil)
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		b.Write(line.Value(source))
	}
	y := b.Bytes()

	rm := Reminder{}
	if err := yaml.UnmarshalStrict(y, &rm); err != nil {
		return ast.WalkStop, err
	}

	if err := tmpl.Execute(writer, &rm); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}