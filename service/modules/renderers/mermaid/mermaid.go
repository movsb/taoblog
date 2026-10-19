// Package mermaid renders Mermaid code blocks to SVG on the server using Pele.
package mermaid

//go:generate curl -fL https://cdn.jsdelivr.net/npm/pele@1.6.0/dist/pele.global.js -o pele.global.js

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dop251/goja"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/yuin/goldmark/parser"
)

//go:embed pele.global.js style1.css
var assets embed.FS

func init() {
	dynamic.RegisterInit(func() {
		dynamic.WithRoots(`mermaid`, nil, nil, assets, utils.NewOSDirFS(string(dir.SourceAbsoluteDir())))
		dynamic.WithStyles(`mermaid`, `style1.css`)
	})
}

type Mermaid struct {
	disabled bool
}

type Option func(*Mermaid)

func WithDisabled(disabled bool) Option {
	return func(m *Mermaid) { m.disabled = disabled }
}

func New(options ...Option) *Mermaid {
	m := &Mermaid{}
	for _, option := range options {
		option(m)
	}
	return m
}

func (m *Mermaid) RenderFencedCodeBlock(w io.Writer, _ string, _ parser.Attributes, source []byte) error {
	// Return an error rather than displaying one, so comment validation rejects
	// the source and it cannot become active content when Mermaid is re-enabled.
	if m.disabled {
		return errors.New("Mermaid 渲染失败：评论中暂不允许使用 Mermaid 图表（仅管理员可用）")
	}
	svg, err := render(string(source))
	if err != nil {
		gold_utils.RenderError(w, err)
		return nil
	}
	_, err = io.WriteString(w, `<div class="mermaid">`+svg+`</div>`)
	return err
}

type runtime struct {
	mu sync.Mutex
	vm *goja.Runtime
	fn goja.Callable
}

var loadRuntime = sync.OnceValues(func() (*runtime, error) {
	vm := goja.New()
	script, err := assets.ReadFile(`pele.global.js`)
	if err != nil {
		return nil, err
	}
	if _, err = vm.RunString(string(script)); err != nil {
		return nil, err
	}
	fn, ok := goja.AssertFunction(vm.Get(`pele`).ToObject(vm).Get(`render`))
	if !ok {
		return nil, fmt.Errorf("Pele render function unavailable")
	}
	return &runtime{vm: vm, fn: fn}, nil
})

var diagramID atomic.Uint64

func render(source string) (string, error) {
	rt, err := loadRuntime()
	if err != nil {
		return "", err
	}
	rt.mu.Lock()
	defer rt.mu.Unlock()

	// Goja permits Interrupt from another goroutine. Wait for the timer callback
	// before clearing it, so a late timeout cannot interrupt the next diagram.
	done := make(chan struct{})
	timer := time.AfterFunc(3*time.Second, func() {
		rt.vm.Interrupt("Mermaid rendering timed out")
		close(done)
	})
	defer func() {
		if !timer.Stop() {
			<-done
		}
		rt.vm.ClearInterrupt()
	}()

	options := rt.vm.ToValue(map[string]any{
		"idPrefix": fmt.Sprintf("mermaid-%d-", diagramID.Add(1)),
	})
	result, err := rt.fn(goja.Undefined(), rt.vm.ToValue(source), options)
	if err != nil {
		return "", err
	}
	return result.ToObject(rt.vm).Get("svg").String(), nil
}
