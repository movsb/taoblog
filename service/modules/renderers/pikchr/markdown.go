package pikchr

import (
	"embed"
	"io"

	"github.com/gopikchr/gopikchr"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/theme/modules/sass"
	"github.com/yuin/goldmark/parser"
)

//go:generate sass --no-source-map style.scss style.css

//go:embed style.css
var _root embed.FS

func init() {
	dynamic.RegisterInit(func() {
		dynamic.Dynamic[`pikchr`] = dynamic.Content{
			Styles: []string{
				string(utils.Must1(_root.ReadFile(`style.css`))),
			},
		}
		sass.WatchDefaultAsync(string(dir.SourceAbsoluteDir()))
	})
}

type PikchrOption func(*Pikchr)

type Pikchr struct{}

func New(options ...PikchrOption) *Pikchr {
	p := &Pikchr{}

	for _, opt := range options {
		opt(p)
	}

	return p
}

func (p *Pikchr) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) (outErr error) {
	defer utils.CatchAsError(&outErr)

	input := string(source)

	const (
		class1 = `pikchr`
		class2 = `pikchr dark`
	)

	var (
		output1, width1, height1, err1 = gopikchr.Convert(input, gopikchr.WithSVGClass(class1))
		output2, width2, height2, _    = gopikchr.Convert(input, gopikchr.WithSVGClass(class2), gopikchr.WithDarkMode())
	)

	if err1 != nil {
		// output* 里面会包含具体的错误，不需要返回错误。
		// return fmt.Errorf(`渲染失败：%w`, errors.Join(err1, err2))
		w.Write([]byte(output1))
		return nil
	}

	w.Write([]byte(output1))
	w.Write([]byte(output2))

	_, _, _, _ = width1, width2, height1, height2

	return nil
}
