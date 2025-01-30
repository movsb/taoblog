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
