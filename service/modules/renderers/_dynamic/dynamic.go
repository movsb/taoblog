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

var Dynamic = map[string]Content{}

var once sync.Once
var style string
var script string
var files []fs.FS
var mod = time.Now()

type Handler struct {
}

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

	path := r.URL.Path[1:]

	switch path {
	case `style`:
		http.ServeContent(w, r, `style.css`, mod, strings.NewReader(style))
		return
	case `script`:
		http.ServeContent(w, r, `script.js`, mod, strings.NewReader(script))
		return
	}

	for _, f := range files {
		p, err := f.Open(path)
		if err == nil {
			p.Close()
			http.ServeFileFS(w, r, f, path)
			return
		}
	}

	http.NotFound(w, r)
}
