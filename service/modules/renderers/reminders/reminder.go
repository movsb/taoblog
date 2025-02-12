package reminders

import (
	"bytes"
	"embed"
	"html/template"

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

	// TODO 在 Transform 的时候实现，以实现不渲染获取到数据。
	if r.out != nil {
		*r.out = append(*r.out, &rm)
	}

	if err := tmpl.Execute(writer, &rm); err != nil {
		return ast.WalkStop, err
	}

	return ast.WalkContinue, nil
}
