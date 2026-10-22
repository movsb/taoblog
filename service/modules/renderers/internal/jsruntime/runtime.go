// Package jsruntime serializes access to moejs and owns its call data.
package jsruntime

import (
	"context"
	"fmt"
	"sync"

	"github.com/Calcium-Ion/moejs"
	"github.com/Calcium-Ion/moejs/eventloop"
)

type Argument struct {
	Name  string
	Value any
}
type Runtime struct {
	mu   sync.Mutex
	vm   *moejs.Runtime
	loop *eventloop.Loop
}

func New(ctx context.Context, timers bool, libs ...[]byte) (*Runtime, error) {
	r := &Runtime{vm: moejs.NewRuntime(moejs.Options{})}
	if timers {
		var err error
		r.loop, err = eventloop.New(r.vm, eventloop.Options{})
		if err != nil {
			return nil, err
		}
	}
	for i, lib := range libs {
		script, err := moejs.CompileScript(fmt.Sprintf("library-%d.js", i), string(lib))
		if err != nil {
			return nil, err
		}
		if err = r.do(ctx, func() error { _, err := r.vm.RunScript(script); return err }); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Start cancellation inside the dispatched task so an early interrupt cannot
// stop the loop before the task signals its completion.
func (r *Runtime) do(ctx context.Context, fn func() error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	var err error
	call := func() {
		done := make(chan struct{})
		stop := context.AfterFunc(ctx, func() { r.vm.Interrupt(context.Cause(ctx)); close(done) })
		defer func() {
			if !stop() {
				<-done
			}
			r.vm.ClearInterrupt()
			r.vm.ReleaseCallData()
		}()
		err = fn()
	}
	if r.loop == nil {
		call()
	} else {
		if err := r.loop.Start(); err != nil {
			return err
		}
		done := make(chan struct{})
		if !r.loop.RunOnLoop(func(*moejs.Runtime) { defer close(done); call() }) {
			_ = r.loop.Stop()
			return fmt.Errorf("JavaScript event loop unavailable")
		}
		<-done
		// Stop before reuse, including after interrupts. Pending timers survive.
		loopErr := r.loop.Stop()
		if err == nil {
			err = loopErr
		}
		r.vm.ClearInterrupt()
	}
	return err
}

func (r *Runtime) Execute(ctx context.Context, source string, args ...Argument) (string, error) {
	script, err := moejs.CompileScript("render.js", source)
	if err != nil {
		return "", err
	}
	return r.Run(ctx, script, args...)
}

// Convert results on the owning goroutine before releasing call data.
func (r *Runtime) Run(ctx context.Context, script *moejs.Script, args ...Argument) (string, error) {
	var out string
	err := r.do(ctx, func() error {
		defer func() {
			realm := r.vm.Realm()
			for _, arg := range args {
				realm.Global.Delete(realm, realm.KeyFromGoString(arg.Name))
			}
		}()
		for _, arg := range args {
			if err := r.vm.SetGlobal(arg.Name, arg.Value); err != nil {
				return err
			}
		}
		value, err := r.vm.RunScript(script)
		if err != nil {
			return err
		}
		result, err := r.vm.ToGo(value)
		if err != nil {
			return err
		}
		var ok bool
		out, ok = result.(string)
		if !ok {
			return fmt.Errorf("JavaScript renderer returned %T, expected string", result)
		}
		return nil
	})
	return out, err
}
