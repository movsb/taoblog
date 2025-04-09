package echarts

import (
	"context"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"
	"github.com/movsb/taoblog/modules/utils"
)

// JavaScript 运行时，必要时移出去作为公共模块。
type Runtime struct {
	ev *eventloop.EventLoop
}

func NewRuntime(ctx context.Context, libs ...[]byte) (_ *Runtime, outErr error) {
	defer utils.CatchAsError(&outErr)

	ev := eventloop.NewEventLoop()
	ev.Start()

	run(ev, func(r *goja.Runtime) {
		for _, lib := range libs {
			utils.Must1(r.RunString(string(lib)))
		}
	})

	return &Runtime{ev: ev}, nil

}

func run(ev *eventloop.EventLoop, fn func(r *goja.Runtime)) {
	wait := make(chan struct{})
	ev.RunOnLoop(func(r *goja.Runtime) {
		defer close(wait)
		fn(r)
	})
	<-wait
}

type Argument struct {
	Name  string
	Value any
}

// arguments 被设置到全局，执行完成后删除。
func (r *Runtime) Execute(ctx context.Context, script string, arguments ...Argument) (_ goja.Value, outErr error) {
	defer utils.CatchAsError(&outErr)

	var val goja.Value
	var err error

	run(r.ev, func(r *goja.Runtime) {
		for _, arg := range arguments {
			utils.Must(r.Set(arg.Name, arg.Value))
		}

		defer func() {
			for _, arg := range arguments {
				r.GlobalObject().Delete(arg.Name)
			}
		}()
		val, err = r.RunString(script)
	})

	return val, err
}
