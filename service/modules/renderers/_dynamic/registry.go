package dynamic

import (
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"reflect"
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
	Styles  [][]byte
	Scripts [][]byte
	Root    fs.FS
}

var (
	_Dynamic   = map[string]*_Content{}
	once       sync.Once
	roots      *http.ServeMux
	mod        = time.Now()
	reloadAll  atomic.Bool
	reloadLock sync.RWMutex
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

func WithRoot(module string, embed, root fs.FS) {
	initModule(module).Root = utils.IIF(version.DevMode(), root, embed)
}

func WithStyles(module string, embed, root fs.FS, paths ...string) {
	c := initModule(module)
	f := utils.IIF(version.DevMode(), root, embed)
	for _, path := range paths {
		content := utils.Must1(fs.ReadFile(f, path))
		c.Styles = append(c.Styles, content)
	}
	if version.DevMode() {
		// 可能是 os.DirFS，但由于是未导出类型，只能自己作判断。
		value := reflect.ValueOf(root)
		if value.Kind() == reflect.String {
			if _, err := fs.Stat(root, `style.scss`); err == nil {
				sass.WatchDefaultAsync(value.String())
				go func() {
					n := utils.NewDirFSWithNotify(value.String()).(utils.FsWithChangeNotify)
					for e := range n.Changed() {
						name := strings.TrimPrefix(strings.TrimPrefix(e.Name, value.String()), `/`)
						if slices.Contains(paths, name) && e.Has(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) {
							// log.Println(`需要重新加载样式`, e)
							reloadAll.Store(true)
						}
					}
				}()
			}
		}
	}
}

func WithScripts(module string, embed, root fs.FS, paths ...string) {
	c := initModule(module)
	f := utils.IIF(version.DevMode(), root, embed)
	for _, path := range paths {
		content := utils.Must1(fs.ReadFile(f, path))
		c.Scripts = append(c.Scripts, content)
	}
	if version.DevMode() {
		// 可能是 os.DirFS，但由于是未导出类型，只能自己作判断。
		value := reflect.ValueOf(root)
		if value.Kind() == reflect.String {
			go func() {
				n := utils.NewDirFSWithNotify(value.String()).(utils.FsWithChangeNotify)
				for e := range n.Changed() {
					name := strings.TrimPrefix(strings.TrimPrefix(e.Name, value.String()), `/`)
					if slices.Contains(paths, name) && e.Has(fsnotify.Create|fsnotify.Write|fsnotify.Rename|fsnotify.Remove) {
						log.Println(`需要重新加载脚本`, e)
						reloadAll.Store(true)
					}
				}
			}()
		}
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
				if strings.HasSuffix(r.URL.Path, `/`) || utils.Must1(f.Stat()).IsDir() {
					http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
					return
				}
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
