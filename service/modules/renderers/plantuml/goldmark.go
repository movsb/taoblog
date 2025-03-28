package plantuml

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	dynamic "github.com/movsb/taoblog/service/modules/renderers/_dynamic"
	"github.com/yuin/goldmark/parser"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

func init() {
	dynamic.RegisterInit(func() {
		const module = `plantuml`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(`plantuml`, `style.css`)
	})
}

type _PlantUMLRenderer struct {
	server string // 可以是 api 前缀
	format string

	cache func(key string, loader func(context.Context) ([]byte, error)) ([]byte, error)
}

func New(server string, format string, options ...Option) *_PlantUMLRenderer {
	p := &_PlantUMLRenderer{
		server: server,
		format: format,
	}
	for _, opt := range options {
		opt(p)
	}

	if p.cache == nil {
		p.cache = func(key string, loader func(ctx context.Context) ([]byte, error)) ([]byte, error) {
			return loader(context.Background())
		}
	}

	return p
}

// 错误被忽略了，因为返回的渲染结果包含错误。
func (p *_PlantUMLRenderer) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) error {
	compressed, err := compress(source)
	if err != nil {
		p.error(w)
		log.Println(`渲染失败`, err)
		return nil
	}

	got, err := p.cache(compressed, func(ctx context.Context) ([]byte, error) {
		light, dark, err := p.fetch(ctx, compressed)
		if err != nil {
			return nil, err
		}
		log.Println(`not using cache for plantuml ...`)
		buf := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buf).Encode(_Cache{Light: light, Dark: dark}); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	})
	if err != nil {
		p.error(w)
		log.Println(`渲染失败`, err)
		return nil
	}

	var cache _Cache
	if err := json.NewDecoder(bytes.NewReader(got)).Decode(&cache); err != nil {
		return err
	}

	w.Write(cache.Light)
	w.Write(cache.Dark)

	return nil
}

// TODO fallback 到用链接。
func (p *_PlantUMLRenderer) error(w io.Writer) {
	fmt.Fprintln(w, `<p style="color:red">PlantUML 渲染失败。</p>`)
}

func (p *_PlantUMLRenderer) fetch(ctx context.Context, compressed string) ([]byte, []byte, error) {
	var (
		content1, content2 []byte
		err1, err2         error
	)

	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	wg.Add(2)
	go func() {
		defer wg.Done()
		content1, err1 = fetch(ctx, p.server, p.format, compressed, false)
	}()
	go func() {
		defer wg.Done()
		content2, err2 = fetch(ctx, p.server, p.format, compressed, true)
	}()
	wg.Wait()

	// 全部错误才算错。
	if err1 != nil && err2 != nil {
		return nil, nil, errors.Join(err1, err2)
	}

	if len(content1) > 0 {
		content1 = style(content1, false)
	}
	if len(content2) > 0 {
		content2 = style(content2, true)
	}

	content1 = strip(content1)
	content2 = strip(content2)

	return content1, content2, nil
}
