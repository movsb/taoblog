package alerts

import (
	"embed"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

//go:embed assets style.css
var _embed embed.FS
var _static = os.DirFS(dir.SourceAbsoluteDir().Join())

//go:generate sass --style compressed --no-source-map style.scss style.css

func init() {
	dynamic.RegisterInit(func() {
		const module = `alerts`
		dynamic.WithRoots(module, nil, nil, _embed, _static)
		dynamic.WithStyles(module, `style.css`)
	})
}

type Alerts struct{}

func New() *Alerts {
	return &Alerts{}
}

var _ interface {
	goldmark.Extender
	renderer.NodeRenderer
	parser.ParagraphTransformer
} = (*Alerts)(nil)

func (e *Alerts) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithParagraphTransformers(util.Prioritized(e, 100)))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(util.Prioritized(e, 100)))
}

var re = regexp.MustCompile(`^\[!((?i:note|tip|important|warning|caution))]$`)

var kind = ast.NewNodeKind(`alert`)

type Alert struct {
	ast.BaseBlock

	text string
}

func (n *Alert) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}
func (n *Alert) Kind() ast.NodeKind {
	return kind
}

func (e *Alerts) RegisterFuncs(r renderer.NodeRendererFuncRegisterer) {
	r.Register(kind, e.RenderAlert)
}

func (e *Alerts) Transform(node *ast.Paragraph, reader text.Reader, pc parser.Context) {
	if p := node.Parent(); p != nil {
		if _, isBlock := p.(*ast.Blockquote); !isBlock {
			return
		}
	}
	lines := node.Lines()
	firstLine := reader.Value(lines.At(0))
	// fast path.
	if !(len(firstLine) > 3 && firstLine[0] == '[') {
		return
	}
	if p := len(firstLine) - 1; firstLine[p] == '\n' {
		firstLine = firstLine[:p]
	}
	match := re.FindSubmatchIndex(firstLine)
	if match == nil {
		return
	}
	t := firstLine[match[2]:match[3]]
	a := Alert{text: string(t)}
	node.Parent().InsertBefore(node.Parent(), node, &a)
	// gold_utils.AddClass(node.Parent(), `alert alert-`+strings.ToLower(a.text))
	segments := text.NewSegments()
	segments.AppendAll(lines.Sliced(1, lines.Len()))
	node.SetLines(segments)
}

func (e *Alerts) RenderAlert(writer util.BufWriter, source []byte, n ast.Node, entering bool) (ast.WalkStatus, error) {
	a := n.(*Alert)
	if entering {
		svg, _ := _embed.ReadFile(fmt.Sprintf(`assets/%s.svg`, strings.ToLower(a.text)))
		fmt.Fprintf(writer, `<p class="alert alert-%s">%s%s`, strings.ToLower(a.text), svg, strings.Title(a.text))
		return ast.WalkSkipChildren, nil
	} else {
		fmt.Fprintln(writer, `</p>`)
		return ast.WalkSkipChildren, nil
	}
}
