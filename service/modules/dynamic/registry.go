package dynamic

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/theme/modules/sass"
)

type _Content struct {
	styleFiles  []string
	scriptFiles []string

	private fs.FS
	public  fs.FS

	handler http.Handler
}

var (
	_Dynamic    = map[string]*_Content{}
	once        sync.Once
	roots       *http.ServeMux
	Mod         time.Time
	reloadAll   atomic.Bool
	reloadLock  sync.RWMutex
	_invalidate func()
)

type ContentChanged interface {
	Reload() <-chan struct{}
}

func initModule(module string) *_Content {
	c := _Dynamic[module]
	if c != nil {
		return c
	}
	c = &_Content{}
	_Dynamic[module] = c
	return c
}

func WithRoots(module string, publicEmbed, publicRoot, privateEmbed, privateRoot fs.FS) {
	c := initModule(module)
	c.private = utils.IIF(version.DevMode(), privateRoot, privateEmbed)
	c.public = utils.IIF(version.DevMode(), publicRoot, publicEmbed)
}

func filesExists(f fs.FS, paths ...string) {
	for _, p := range paths {
		f := utils.Must1(f.Open(p))
		f.Close()
	}
}

func WithStyles(module string, paths ...string) {
	c := initModule(module)
	c.styleFiles = paths

	filesExists(c.private, paths...)

	if dirFS, ok := c.private.(*utils.OSDirFS); ok {
		if _, err := fs.Stat(c.private, `style.scss`); err == nil {
			root := path.Clean(dirFS.Root())
			if version.DevMode() {
				sass.WatchDefaultAsync(root)
			}
			go func() {
				events, close := utils.Must2(dirFS.Watch())
				defer close()
				for e := range events {
					name := strings.TrimPrefix(strings.TrimPrefix(e.Name, root), `/`)
					if slices.Contains(paths, name) && e.Has(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) {
						// log.Println(`需要重新加载样式`, e)
						reloadAll.Store(true)
					}
				}
			}()
		}
	}

	if cc, ok := c.private.(ContentChanged); ok {
		go func() {
			for range cc.Reload() {
				// log.Println(`需要重新加载样式`, module)
				reloadAll.Store(true)
			}
		}()
	}
}

func WithScripts(module string, paths ...string) {
	c := initModule(module)
	c.scriptFiles = paths

	filesExists(c.private, paths...)

	if dirFS, ok := c.private.(*utils.OSDirFS); ok {
		root := path.Clean(dirFS.Root())
		go func() {
			events, close := utils.Must2(dirFS.Watch())
			defer close()
			for e := range events {
				name := strings.TrimPrefix(strings.TrimPrefix(e.Name, root), `/`)
				if slices.Contains(paths, name) && e.Has(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) {
					log.Println(`需要重新加载脚本`, e)
					reloadAll.Store(true)
				}
			}
		}()
	}

	if cc, ok := c.private.(ContentChanged); ok {
		go func() {
			for range cc.Reload() {
				log.Println(`需要重新加载脚本`, module)
				reloadAll.Store(true)
			}
		}()
	}
}

// 线程安全。
func WithHandler(module string, handler http.Handler) {
	reloadLock.Lock()
	defer reloadLock.Unlock()

	c := initModule(module)
	c.handler = handler

	reloadAll.Store(true)
}

func initContents() {
	var (
		styleBuilder  bytes.Buffer
		scriptBuilder bytes.Buffer
	)

	Mod = time.Now()
	roots = http.NewServeMux()
	defer _invalidate()

	for module, d := range _Dynamic {
		if d.public != nil || d.handler != nil {
			handler := func(w http.ResponseWriter, r *http.Request) {
				if d.public != nil {
					// 不直接 ServeFS 是因为 embed.FS 不支持 ModTime.
					// 进而导致浏览器缓存不生效。
					f, err := http.FS(d.public).Open(r.URL.Path)
					if err == nil {
						defer f.Close()
						if strings.HasSuffix(r.URL.Path, `/`) || utils.Must1(f.Stat()).IsDir() {
							http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
							return
						}
						http.ServeContent(w, r, r.URL.Path, Mod, f)
						return
					}
					// TODO 其它错误被忽略了。
				}
				if d.handler != nil {
					d.handler.ServeHTTP(w, r)
				} else {
					http.NotFound(w, r)
				}
			}
			roots.Handle(
				fmt.Sprintf(`GET /%s/`, module),
				http.StripPrefix(`/`+module, http.HandlerFunc(handler)),
			)
		}

		read := func(path string) []byte {
			return utils.Must1(fs.ReadFile(d.private, path))
		}

		fmt.Fprintf(&styleBuilder, "/* %s */\n", module)
		for _, s := range d.styleFiles {
			styleBuilder.Write(read(s))
			styleBuilder.WriteByte('\n')
		}
		styleBuilder.WriteByte('\n')

		fmt.Fprintf(&scriptBuilder, "// %s\n", module)
		for _, s := range d.scriptFiles {
			scriptBuilder.Write(read(s))
			scriptBuilder.WriteByte('\n')
		}
		scriptBuilder.WriteByte('\n')
	}

	style := bytes.NewReader(styleBuilder.Bytes())
	script := bytes.NewReader(scriptBuilder.Bytes())

	roots.HandleFunc(`GET /style.css`, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(`Cache-Control`, `private, no-cache, must-revalidate`)
		http.ServeContent(w, r, `style.css`, Mod, style)
	})
	roots.HandleFunc(`GET /script.js`, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(`Cache-Control`, `private, no-cache, must-revalidate`)
		http.ServeContent(w, r, `script.js`, Mod, script)
	})
}

////////////////////////////////////////////////////////////////////////////////

var (
	inits       []func()
	onceInits   sync.Once
	initsCalled atomic.Bool
)

func RegisterInit(init func()) {
	if initsCalled.Load() {
		init()
		reloadAll.Store(true)
		return
	}
	inits = append(inits, init)
}

func callInits() {
	for _, init := range inits {
		init()
	}
	initsCalled.Store(true)
}

// 主题、扩展有变动的时候可以调用此方法更新动态内容。
func Reload() {
	reloadAll.Store(true)
	initAll()
}
