package dynamic

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
)

type _Content struct {
	Styles  [][]byte
	Scripts [][]byte
	Root    fs.FS
}

var (
	_Dynamic = map[string]*_Content{}
	once     sync.Once
	roots    *http.ServeMux
	mod      = time.Now()
)

func initModule(module string) *_Content {
	c := _Dynamic[module]
	if c != nil {
		return c
	}
	c = &_Content{}
	_Dynamic[module] = c
	return c
}

func WithRoot(module string, fsys fs.FS) {
	initModule(module).Root = fsys
}

func WithStyles(module string, fsys fs.FS, paths ...string) {
	c := initModule(module)
	for _, path := range paths {
		content := utils.Must1(fs.ReadFile(fsys, path))
		c.Styles = append(c.Styles, content)
	}
}

func WithScripts(module string, fsys fs.FS, paths ...string) {
	c := initModule(module)
	for _, path := range paths {
		content := utils.Must1(fs.ReadFile(fsys, path))
		c.Scripts = append(c.Scripts, content)
	}
}

func initContents() {
	var (
		styleBuilder  bytes.Buffer
		scriptBuilder bytes.Buffer
	)

	roots = http.NewServeMux()

	for module, d := range _Dynamic {
		if d.Root != nil {
			handler := func(w http.ResponseWriter, r *http.Request) {
				// 不直接 ServeFS 是因为 embed.FS 不支持 ModTime.
				// 进而导致浏览器缓存不生效。
				f := utils.Must1(http.FS(d.Root).Open(r.URL.Path))
				defer f.Close()
				http.ServeContent(w, r, r.URL.Path, mod, f)
			}
			roots.Handle(
				fmt.Sprintf(`GET /%s/`, module),
				http.StripPrefix(`/`+module, http.HandlerFunc(handler)),
			)
		}

		fmt.Fprintf(&styleBuilder, "/* %s */\n", module)
		for _, s := range d.Styles {
			styleBuilder.Write(s)
			styleBuilder.WriteByte('\n')
		}
		styleBuilder.WriteByte('\n')

		fmt.Fprintf(&scriptBuilder, "// %s\n", module)
		for _, s := range d.Scripts {
			scriptBuilder.Write(s)
			scriptBuilder.WriteByte('\n')
		}
		scriptBuilder.WriteByte('\n')
	}

	style := bytes.NewReader(styleBuilder.Bytes())
	script := bytes.NewReader(scriptBuilder.Bytes())

	roots.HandleFunc(`GET /style.css`, func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, `style.css`, mod, style)
	})
	roots.HandleFunc(`GET /script.js`, func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, `script.js`, mod, script)
	})
}

////////////////////////////////////////////////////////////////////////////////

var (
	inits     []func()
	onceInits sync.Once
)

func RegisterInit(init func()) {
	inits = append(inits, init)
}

func callInits() {
	for _, init := range inits {
		init()
	}
}
