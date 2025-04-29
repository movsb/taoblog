package dynamic_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/movsb/taoblog/service/modules/renderers/emojis"
)

func TestModTime(t *testing.T) {
	s := httptest.NewServer(http.StripPrefix(dynamic.Prefix, dynamic.New(func() {})))
	defer s.Close()

	baseURL, _ := url.Parse(s.URL)
	dogeURL := baseURL.JoinPath(emojis.BaseURLForDynamic.JoinPath(`weixin/doge.png`).String())
	rsp := utils.Must1(http.Get(dogeURL.String()))
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		t.Fatal(rsp.Status)
	}

	if rsp.Header.Get(`Last-Modified`) == `` {
		t.Fatal(`no last-modified header`)
	}
}
