package goldmark_katex

import (
	"context"

	"github.com/dop251/goja"
	"github.com/movsb/taoblog/modules/utils"
)

// JavaScript 运行时，必要时移出去作为公共模块。
type Runtime struct {
	vm *goja.Runtime
}

func NewRuntime(ctx context.Context, libs ...[]byte) (_ *Runtime, outErr error) {
	defer utils.CatchAsError(&outErr)

	vm := goja.New()

	for _, lib := range libs {
		utils.Must1(vm.RunString(string(lib)))
	}

	return &Runtime{vm: vm}, nil
}

type Argument struct {
	Name  string
	Value any
}

// arguments 被设置到全局，执行完成后删除。
func (r *Runtime) Execute(ctx context.Context, script string, arguments ...Argument) (_ goja.Value, outErr error) {
	defer utils.CatchAsError(&outErr)

	for _, arg := range arguments {
		utils.Must(r.vm.Set(arg.Name, arg.Value))
	}

	defer func() {
		for _, arg := range arguments {
			r.vm.GlobalObject().Delete(arg.Name)
		}
	}()

	return r.vm.RunString(script)
}
