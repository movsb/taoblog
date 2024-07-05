package e2e_test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func expectHTTPGetWithStatusCode(relativeURL string, code int) {
	u := utils.Must1(url.Parse(`http://` + Server.HTTPAddr))
	ur := utils.Must1(url.Parse(relativeURL))
	urlFinal := u.ResolveReference(ur)
	rsp, err := http.Get(urlFinal.String())
	if err != nil {
		panic(err)
	}
	if rsp.StatusCode != code {
		panic(fmt.Sprintf(`状态码不相等：%d with %d`, rsp.StatusCode, code))
	}
}

func ExpectError(t *testing.T, err error) {
	if err == nil {
		panic(`期待错误，但是实际没有。`)
	}
}
