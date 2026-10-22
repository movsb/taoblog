// Package mermaid renders Mermaid code blocks to SVG on the server using Pele.
package mermaid

//go:generate curl -fL https://cdn.jsdelivr.net/npm/pele@1.6.0/dist/pele.global.js -o pele.global.js

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Calcium-Ion/moejs"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/renderers/internal/jsruntime"
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
	vm     *jsruntime.Runtime
	script *moejs.Script
}

var loadRuntime = sync.OnceValues(func() (*runtime, error) {
	lib, err := assets.ReadFile(`pele.global.js`)
	if err != nil {
		return nil, err
	}
	vm, err := jsruntime.New(context.Background(), false, lib)
	if err != nil {
		return nil, err
	}
	script, err := moejs.CompileScript("pele-render.js", `pele.render(source, options).svg`)
	if err != nil {
		return nil, err
	}
	return &runtime{vm: vm, script: script}, nil
})

var diagramID atomic.Uint64

func render(source string) (string, error) {
	rt, err := loadRuntime()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return rt.vm.Run(ctx, rt.script,
		jsruntime.Argument{Name: "source", Value: source},
		jsruntime.Argument{Name: "options", Value: map[string]any{"idPrefix": fmt.Sprintf("mermaid-%d-", diagramID.Add(1))}},
	)
}
