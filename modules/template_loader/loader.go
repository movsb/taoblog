package templateloader

import (
	"context"
	"html/template"
	fspkg "io/fs"
	"log"
	"path/filepath"
	"sync/atomic"
	"time"
)

type TemplateLoader struct {
	name string

	t atomic.Value

	fs       fspkg.FS // 从二进制文件 embedded 资源加载
	root     string   // 从本地文件系统加载
	patterns []string // 待加载的资源列表

	// 如果使用本地文件系统加载，间隔多久进行重新加载，方便调试模板文件。
	// 如果不设置，则不自动重新加载。
	reloadInterval time.Duration

	// 当前使用哪种模式。
	useLocal bool
}

type Option func(*TemplateLoader)

func WithPatterns(patterns ...string) Option {
	return func(t *TemplateLoader) {
		t.patterns = patterns
	}
}

func WithEmbeddedFS(fs fspkg.FS) Option {
	return func(t *TemplateLoader) {
		t.fs = fs
	}
}

func WithLocalFS(root string) Option {
	return func(t *TemplateLoader) {
		t.root = root
	}
}

func WithEnableLocalFS(enable bool) Option {
	return func(t *TemplateLoader) {
		t.useLocal = enable
	}
}

func WithReloadInterval(interval time.Duration) Option {
	return func(t *TemplateLoader) {
		t.reloadInterval = interval
	}
}

func NewTemplateLoader(ctx context.Context, name string, options ...Option) *TemplateLoader {
	t := &TemplateLoader{
		name: name,
	}

	t.t.Store(template.New(name))

	for _, option := range options {
		option(t)
	}

	if t.fs == nil && t.root == "" {
		panic(`both fs and root are nil`)
	}
	if len(t.patterns) == 0 {
		panic(`no patterns specified`)
	}

	// 使用本地资源时允许加载错误的资源。
	t2, err := t.load()
	if err != nil {
		if !t.useLocal {
			panic(err)
		}
		// err 为为空时 t2 是个空模板，可以使用。
	}
	t.t.Store(t2)

	if t.useLocal {
		go t.run(ctx)
	}

	return t
}

func (t *TemplateLoader) T() *template.Template {
	return t.t.Load().(*template.Template)
}

func (t *TemplateLoader) run(ctx context.Context) {
	if !t.useLocal {
		panic("not local template loader")
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(t.reloadInterval):
			t2, err := t.load()
			if err != nil {
				log.Println(err)
				continue
			}
			t.t.Store(t2)
		}
	}
}

// 加载失败时，第一个参数也不为空。
func (t *TemplateLoader) load() (*template.Template, error) {
	var err error

	t2 := template.New(t.name)

	if t.useLocal {
		for _, pattern := range t.patterns {
			fullPattern := filepath.Join(t.root, pattern)
			t2, err = t2.ParseGlob(fullPattern)
			if err != nil {
				return t2, err
			}
		}
	} else {
		t2, err = t2.ParseFS(t.fs, t.patterns...)
		if err != nil {
			return t2, err
		}
	}

	return t2, nil
}
