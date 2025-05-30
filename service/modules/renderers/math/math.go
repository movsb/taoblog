package math

import (
	"context"
	"embed"
	"io/fs"
	"sync"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	goldmark_katex "github.com/movsb/taoblog/service/modules/renderers/math/goldmark"
	"github.com/phuslu/lru"
	"github.com/yuin/goldmark"
)

//go:generate sass --style compressed --no-source-map katex/style.scss katex/style.css

//go:embed static katex/katex.min.stripped.css katex/style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

func init() {
	dynamic.RegisterInit(func() {
		const module = `katex`
		katexDirRoot := utils.Must1(fs.Sub(_root, `katex`))
		katexDirEmbed := utils.Must1(fs.Sub(_embed, `katex`))
		staticDirRoot := utils.Must1(fs.Sub(_root, `static`))
		staticDirEmbed := utils.Must1(fs.Sub(_embed, `static`))
		dynamic.WithRoots(module, staticDirEmbed, staticDirRoot, katexDirEmbed, katexDirRoot)
		dynamic.WithStyles(module, `katex.min.stripped.css`, `style.css`)
	})
}

type Math struct {
	goldmark.Extender
}

var (
	once  sync.Once
	cache *lru.TTLCache[goldmark_katex.CacheKey, []byte]
	rt    *goldmark_katex.Runtime
)

var (
	//go:embed katex/katex.min.js
	_katexBinary []byte
	//go:embed katex/mhchem.min.js
	_chemBinary []byte
)

func New() goldmark.Extender {
	once.Do(func() {
		cache = lru.NewTTLCache[goldmark_katex.CacheKey, []byte](128)
		rt = utils.Must1(goldmark_katex.NewRuntime(context.Background(), _katexBinary, _chemBinary))
	})
	return &Math{
		Extender: goldmark_katex.New(rt, cache),
	}
}
