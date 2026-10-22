package echarts

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRender(t *testing.T) {
	svg, err := render(context.Background(), `let option = {};`, 220, 220)
	if err != nil || !strings.HasPrefix(svg, "<svg ") {
		t.Fatalf("render: %q %v", svg, err)
	}
}

func TestConcurrentRenderAndTimeoutRecovery(t *testing.T) {
	// Initialize before starting the deadline so it measures execution only.
	rt := runtime()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	_, err := rt.Execute(ctx, `for(;;){}`)
	cancel()
	if err == nil {
		t.Fatal("expected timeout")
	}
	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			svg, err := render(context.Background(), `let option = {animation:false,series:[{type:'pie',data:[{name:'中文',value:1}]}]};`, 220, 220)
			if err != nil || !strings.Contains(svg, "中文") {
				t.Errorf("render after timeout: %v", err)
			}
		}()
	}
	wg.Wait()
}
