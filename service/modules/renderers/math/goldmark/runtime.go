package goldmark_katex

import (
	"context"

	"github.com/movsb/taoblog/service/modules/renderers/internal/jsruntime"
)

type Runtime = jsruntime.Runtime
type Argument = jsruntime.Argument

func NewRuntime(ctx context.Context, libs ...[]byte) (*Runtime, error) {
	return jsruntime.New(ctx, false, libs...)
}
