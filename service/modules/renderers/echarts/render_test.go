package echarts

import (
	"context"
	"testing"
)

func TestRender(t *testing.T) {
	t.Log(render(context.Background(), `let option = {};`))
}
