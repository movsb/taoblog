package dynamic

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Content struct {
	Styles  []string
	Scripts []string
	Root    fs.FS
}

// TODO 不要暴露出去，外部通过注册的方式提供。
// 然后，以目录名的方式防止名字冲突。
var Dynamic = map[string]Content{}

var once sync.Once
var style string
var script string
var files []fs.FS
var mod = time.Now()
var once2 sync.Once

func New() http.Handler {
	once2.Do(callInits)
	return &Handler{}
}

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		for name, d := range Dynamic {
			style += fmt.Sprintf("/* %s */\n", name)
			style += strings.Join(d.Styles, "\n")
			style += "\n\n"

			script += fmt.Sprintf("// %s\n", name)
			script += strings.Join(d.Scripts, "\n")
			script += "\n\n"

			if d.Root != nil {
				files = append(files, d.Root)
			}
		}
	})

	path := r.URL.Path

	// TODO 使用 Mux
	switch path {
	case `/style`:
		http.ServeContent(w, r, `style.css`, mod, strings.NewReader(style))
		return
	case `/script`:
		http.ServeContent(w, r, `script.js`, mod, strings.NewReader(script))
		return
	}

	// TODO 使用 OverlayFS
	// TODO 按插件名称组织结构，就无需循环了。
	for _, f := range files {
		// 为了支持 io.Seeker.
		p, err := http.FS(f).Open(path)
		if err == nil {
			defer p.Close()
			// 不使用 http.ServeFileFS：
			// - embed.FS 不支持文件的修改时间
			// - 可以少一次文件打开操作
			http.ServeContent(w, r, path, mod, p)
			return
		}
	}

	http.NotFound(w, r)
}

var (
	inits []func()
)

func RegisterInit(init func()) {
	inits = append(inits, init)
}

func callInits() {
	for _, init := range inits {
		init()
	}
}
