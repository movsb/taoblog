package jsruntime

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestIsolationAndRecovery(t *testing.T) {
	for _, timers := range []bool{false, true} {
		t.Run(fmt.Sprint(timers), func(t *testing.T) {
			r, err := New(context.Background(), timers)
			if err != nil {
				t.Fatal(err)
			}
			if value, err := r.Execute(context.Background(), `options.label + ":" + enabled`, Argument{"options", map[string]any{"label": "中文"}}, Argument{"enabled", true}); err != nil || value != "中文:true" {
				t.Fatalf("binding: %q %v", value, err)
			}
			if value, err := r.Execute(context.Background(), `typeof options + ":" + typeof enabled`); err != nil || value != "undefined:undefined" {
				t.Fatalf("cleanup: %q %v", value, err)
			}
			if _, err := r.Execute(context.Background(), `throw new Error("invalid")`, Argument{"temporary", 42}); err == nil {
				t.Fatal("expected exception")
			}
			if value, err := r.Execute(context.Background(), `typeof temporary`); err != nil || value != "undefined" {
				t.Fatalf("exception cleanup: %q %v", value, err)
			}
			for range 3 {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
				_, err := r.Execute(ctx, `for(;;){}`, Argument{"temporary", 42})
				cancel()
				if err == nil {
					t.Fatal("expected interruption")
				}
				if value, err := r.Execute(context.Background(), `typeof temporary + ":ok"`); err != nil || value != "undefined:ok" {
					t.Fatalf("timeout recovery: %q %v", value, err)
				}
			}
			var wg sync.WaitGroup
			for i := range 12 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					want := fmt.Sprint(i)
					got, err := r.Execute(context.Background(), `String(input)`, Argument{"input", i})
					if err != nil || got != want {
						t.Errorf("concurrent binding: %q want %q, %v", got, want, err)
					}
				}()
			}
			wg.Wait()
		})
	}
}
