package katex

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// WebAssembly 运行时。
//
// 注意：如果出现以下错误：
//
//	panic: wasm error: unreachable
//
// 大概率是代码中抛出了异常（比如：有未识别的函数、变量）。
// 但是目前不知道如何获取到具体的错误。所以请在非 WASI 环境中
// 测试运行好后再交由本运行时运行方能更好地除错。
type WebAssemblyRuntime struct {
	r wazero.Runtime

	wasi api.Closer

	modQuickjs wazero.CompiledModule
	modKatex   wazero.CompiledModule
}

func NewWebAssemblyRuntime(ctx context.Context) (_ *WebAssemblyRuntime, outErr error) {
	defer utils.CatchAsError(&outErr)

	r := WebAssemblyRuntime{
		r: wazero.NewRuntime(ctx),
	}

	r.wasi = utils.Must1(wasi_snapshot_preview1.Instantiate(ctx, r.r))

	r.modQuickjs = utils.Must1(r.r.CompileModule(ctx, utils.Must1(Root.ReadFile(`binary/quickjs.wasm`))))
	r.modKatex = utils.Must1(r.r.CompileModule(ctx, utils.Must1(Root.ReadFile(`binary/katex.wasm`))))

	return &r, nil
}

type KatexArguments struct {
	Tex         string `json:"tex"`
	DisplayMode bool   `json:"displayMode"`
}

func (r *WebAssemblyRuntime) RenderKatex(ctx context.Context, tex string, displayMode bool) (_ string, outErr error) {
	defer utils.CatchAsError(&outErr)

	var (
		input = bytes.NewReader(utils.Must1(json.Marshal(KatexArguments{
			Tex:         tex,
			DisplayMode: displayMode,
		})))
		output = bytes.NewBuffer(nil)
	)

	configBase := wazero.NewModuleConfig().WithStdin(input).WithStdout(output).WithStderr(output).WithStartFunctions()

	modQuickjs := utils.Must1(r.r.InstantiateModule(ctx, r.modQuickjs, configBase.WithName(`javy_quickjs_provider_v3`)))
	defer modQuickjs.Close(ctx)

	modKatex := utils.Must1(r.r.InstantiateModule(ctx, r.modKatex, configBase))
	defer modKatex.Close(ctx)

	utils.Must1(modKatex.ExportedFunction(`_start`).Call(ctx))

	return output.String(), nil
}

func (r *WebAssemblyRuntime) Close() error {
	ctx := context.Background()

	r.modKatex.Close(ctx)
	r.modQuickjs.Close(ctx)

	r.wasi.Close(ctx)
	r.r.Close(ctx)

	return nil
}
