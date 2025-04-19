package rss

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
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

const module = `rss`

func init() {
	dynamic.RegisterInit(func() {
		dynamic.WithRoots(module, nil, nil, _embed, _local)
		dynamic.WithStyles(module, `style.css`)
		dynamic.WithScripts(module, `script.js`)
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

//go:embed post.html style.css script.js
var _embed embed.FS
var _local = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

var tmpl = sync.OnceValue(func() *utils.TemplateLoader {
	return utils.NewTemplateLoader(utils.IIF(version.DevMode(), _local, fs.FS(_embed)), nil, dynamic.Reload)
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

	fmt.Fprintln(w, `<div class="rss-post-list">`)

	for _, d := range data {
		if err := tmpl().GetNamed(`post.html`).Execute(w, d); err != nil {
			log.Println(err)
		}
	}

	fmt.Fprintln(w, `</div>`)

	return nil
}
