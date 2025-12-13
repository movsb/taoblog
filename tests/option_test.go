package e2e_test

import (
	"net/url"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestOption(t *testing.T) {
	r := Serve(t.Context())

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

func TestEnterDebug(t *testing.T) {
	r := Serve(t.Context())

	client := utils.Must1(r.client.Management.EnterDebug(r.guest))
	_, err := client.Recv()
	if err == nil {
		panic(`应该报错`)
	}

	client = utils.Must1(r.client.Management.EnterDebug(r.user1))
	_, err = client.Recv()
	if err == nil {
		panic(`应该报错`)
	}

	client = utils.Must1(r.client.Management.EnterDebug(r.admin))
	rsp, err := client.Recv()
	if err != nil {
		panic(`不应该报错`)
	}
	url := utils.Must1(url.Parse(rsp.Url)).JoinPath(`/pprof/allocs`)
	expectHTTPGetWithStatusCodeAbsoluteURL(r, url.String(), 200)
}
