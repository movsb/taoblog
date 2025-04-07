package e2e_test

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func expectHTTPGetWithStatusCode(r *R, relativeURL string, code int) {
	u := utils.Must1(url.Parse(`http://` + r.server.HTTPAddr()))
	ur := utils.Must1(url.Parse(relativeURL))
	urlFinal := u.ResolveReference(ur)
	rsp, err := http.Get(urlFinal.String())
	if err != nil {
		panic(err)
	}
	if rsp.StatusCode != code {
		_, file, line, _ := runtime.Caller(1)
		io.Copy(os.Stderr, rsp.Body)
		panic(fmt.Sprintf(`[%s:%d] 状态码不相等：got: %d, expect: %d`, file, line, rsp.StatusCode, code))
	}
}

func ExpectError(t *testing.T, err error) {
	if err == nil {
		panic(`期待错误，但是实际没有。`)
	}
}

func expect304(r *R, relativeURL string) {
	u := utils.Must1(url.Parse(`http://` + r.server.HTTPAddr()))
	ur := utils.Must1(url.Parse(relativeURL))
	urlFinal := u.ResolveReference(ur)

	req := utils.Must1(http.NewRequest(http.MethodGet, urlFinal.String(), nil))
	rsp := utils.Must1(http.DefaultClient.Do(req))
	if rsp.StatusCode != 200 {
		panic(`not 200`)
	}
	lastModified := rsp.Header.Get(`Last-Modified`)
	eTag := rsp.Header.Get(`ETag`)

	req = utils.Must1(http.NewRequest(http.MethodGet, urlFinal.String(), nil))
	if lastModified != `` {
		req.Header.Set(`If-Modified-Since`, lastModified)
	}
	if eTag != `` {
		req.Header.Set(`If-Match`, eTag)
	}
	rsp = utils.Must1(http.DefaultClient.Do(req))
	if rsp.StatusCode != 304 {
		panic(`not 304`)
	}
}
