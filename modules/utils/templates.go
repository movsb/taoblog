package utils

import (
	"html/template"
	"io/fs"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type TemplateLoader struct {
	lock sync.RWMutex

	fsys  fs.FS
	funcs template.FuncMap

	partial *template.Template
	named   map[string]*template.Template
}

// refreshed: 初次加载不会调用，后面刷新调用。
func NewTemplateLoader(fsys fs.FS, funcs template.FuncMap, refreshed func()) *TemplateLoader {
	l := TemplateLoader{
		fsys:    fsys,
		funcs:   funcs,
		partial: template.Must(template.New(`partial`).Parse(``)),
		named:   make(map[string]*template.Template),
	}

	bundle := func() {
		l.parsePartial()
		l.parseNamed()
	}

	bundle()

	if watchFS, ok := fsys.(WatchFS); ok {
		log.Println(`Listening for template changes`)
		go func() {
			events, close := Must2(watchFS.Watch())
			defer close()

			debouncer := NewDebouncer(time.Second, func() {
				bundle()
				log.Println(`Re-parsed all partial and named templates`)
				if refreshed != nil {
					refreshed()
				}
			})
			for event := range events {
				if event.Has(fsnotify.Create | fsnotify.Remove | fsnotify.Write) {
					debouncer.Enter()
				}
			}
		}()
	}

	return &l
}

func (l *TemplateLoader) GetNamed(name string) *template.Template {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.named[name]
}
func (l *TemplateLoader) GetPartial(name string) *template.Template {
	l.lock.RLock()
	defer l.lock.RUnlock()
	return l.partial.Lookup(name)
}

func (l *TemplateLoader) parsePartial() {
	t2, err := template.New(`partial`).Funcs(l.funcs).ParseFS(l.fsys, `_*.html`)
	if err != nil {
		if !strings.Contains(err.Error(), `matches no files`) {
			log.Println("\033[31m", err, "\033[m")
			return
		}
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	l.partial = t2
}

func (l *TemplateLoader) parseNamed() {
	l.lock.Lock()
	defer l.lock.Unlock()
	names, _ := fs.Glob(l.fsys, `[^_]*.html`)
	l.named = make(map[string]*template.Template)
	for _, name := range names {
		// NOTE: name 如果包含 pattern 字符的话，这里大概率会出错。奇怪为什么没有按 name parse 的。
		t2, err := template.New(name).Funcs(l.funcs).ParseFS(l.fsys, name)
		if err != nil {
			log.Println("\033[31m", err, "\033[m")
			continue
		}
		l.named[name] = t2
	}
}
