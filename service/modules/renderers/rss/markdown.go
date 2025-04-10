package rss

import (
	"embed"
	"io"
	"io/fs"
	"slices"
	"strings"
	"sync"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/modules/version"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark/parser"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

func init() {
	dynamic.RegisterInit(func() {
		const module = `rss`
		dynamic.WithRoots(module, nil, nil, _embed, _local)
		dynamic.WithStyles(module, `style.css`)
	})
}

type Rss struct {
	task *Task
	post int
}

var _ interface {
	gold_utils.FencedCodeBlockRenderer
} = (*Rss)(nil)

func New(task *Task, postID int) *Rss {
	r := Rss{
		task: task,
		post: postID,
	}
	return &r
}

//go:embed post.html style.css
var _embed embed.FS
var _local = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

var tmpl = sync.OnceValue(func() *utils.TemplateLoader {
	return utils.NewTemplateLoader(utils.IIF(version.DevMode(), _local, fs.FS(_embed)), nil, func() {})
})

func (r *Rss) RenderFencedCodeBlock(w io.Writer, language string, attrs parser.Attributes, source []byte) error {
	urls := strings.Split(string(source), "\n")
	if len(urls) == 1 && urls[0] == `` {
		return nil
	}
	urls = utils.Map(urls, func(u string) string {
		return strings.TrimSpace(u)
	})
	urls = slices.DeleteFunc(urls, func(u string) bool {
		return u == ``
	})

	// TODO 未保存的预览也会影响此结果
	data := r.task.GetLatestPosts(r.post, urls)
	for _, d := range data {
		tmpl().GetNamed(`post.html`).Execute(w, d)
	}

	return nil
}
