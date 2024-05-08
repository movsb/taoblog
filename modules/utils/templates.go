package utils

import (
	"html/template"
	"io/fs"
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type TemplateLoader struct {
	lock sync.RWMutex

	fsys  fs.FS
	funcs template.FuncMap

	partial *template.Template
	named   map[string]*template.Template
}

func NewTemplateLoader(fsys fs.FS, funcs template.FuncMap) *TemplateLoader {
	l := TemplateLoader{
		fsys:    fsys,
		funcs:   funcs,
		partial: template.Must(template.New(`partial`).Parse(``)),
		named:   make(map[string]*template.Template),
	}

	l.parsePartial()
	l.parseNamed()

	if changed, ok := fsys.(FsWithChangeNotify); ok {
		log.Println(`Listening for template changes`)
		go func() {
			for event := range changed.Changed() {
				switch event.Op {
				case fsnotify.Create, fsnotify.Remove, fsnotify.Write:
					l.parsePartial()
					l.parseNamed()
					log.Println(`Re-parsed all partial and named templates`)
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
		log.Println(err)
		return
	}
	l.lock.Lock()
	defer l.lock.Unlock()
	l.partial = t2
}

func (l *TemplateLoader) parseNamed() {
	l.lock.Lock()
	defer l.lock.Unlock()
	names, _ := fs.Glob(l.fsys, `[^_]*.html`)
	for _, name := range names {
		// NOTE: name 如果包含 pattern 字符的话，这里大概率会出错。奇怪为什么没有按 name parse 的。
		t2, err := template.New(name).Funcs(l.funcs).ParseFS(l.fsys, name)
		if err != nil {
			log.Println(err)
			continue
		}
		l.named[name] = t2
	}
}
