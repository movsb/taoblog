package e2e_test

import (
	"context"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestOption(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := Serve(ctx)

	opt := r.server.Main().Options()
	opt.SetString(`s1`, `v1`)
	opt.SetInteger(`i1`, 1)

	if utils.Must1(opt.GetString(`s1`)) != `v1` {
		t.Error(1)
	}
	if utils.Must1(opt.GetStringDefault(`s2`, `v2`)) != `v2` {
		t.Error(2)
	}

	if utils.Must1(opt.GetInteger(`i1`)) != 1 {
		t.Error(1)
	}
	if utils.Must1(opt.GetIntegerDefault(`i2`, 2)) != 2 {
		t.Error(2)
	}
}
