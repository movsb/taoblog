package e2e_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func expectHTTPGetWithStatusCode(r *R, relativeURL string, code int) {
	expectHTTPGetWithStatusCodeAbsoluteURL(r, r.server.JoinPath(relativeURL), code)
}

func expectHTTPGetWithStatusCodeAbsoluteURL(r *R, u string, code int) {
	rsp, err := http.Get(u)
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
		_, file, line, _ := runtime.Caller(1)
		t.Fatal(`期待错误，但是实际没有。`, file, line)
	}
}
func ExpectNoError(t *testing.T, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		t.Fatal(`期待没有错误，但是实际有。`, err, file, line)
	}
}

func expect304(r *R, relativeURL string) {
	u := r.server.JoinPath(relativeURL)

	req := utils.Must1(http.NewRequest(http.MethodGet, u, nil))
	rsp := utils.Must1(http.DefaultClient.Do(req))
	if rsp.StatusCode != 200 {
		panic(`not 200`)
	}
	lastModified := rsp.Header.Get(`Last-Modified`)
	eTag := rsp.Header.Get(`ETag`)

	req = utils.Must1(http.NewRequest(http.MethodGet, u, nil))
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

func TestTaskAuth(t *testing.T) {
	r := Serve(t.Context())

	// 创建流的时候始终不会因为权限报错。

	ExpectError := func(t *testing.T, err error) {
		if err == nil {
			_, file, line, _ := runtime.Caller(2)
			t.Fatal(`期待错误，但是实际没有。`, file, line)
		}
		if st, ok := status.FromError(err); !ok || st.Code() != codes.PermissionDenied {
			t.Fatal(`非为权限错误`)
		}
	}
	ExpectNoError := func(t *testing.T, err error) {
		if err != nil {
			if st, ok := status.FromError(err); !ok || st.Code() != codes.DeadlineExceeded {
				_, file, line, _ := runtime.Caller(2)
				t.Fatal(`期待没有错误，但是实际有。`, err, file, line)
			}
		}
	}
	tc := func(user context.Context, err2 bool) {
		ctx, cancel := context.WithTimeout(user, time.Second*1)
		defer cancel()
		client, err := r.client.Utils.RegisterAutoImageBorderHandler(ctx)
		ExpectNoError(t, err)
		_, err = client.Recv()
		utils.IIF(err2, ExpectError, ExpectNoError)(t, err)
	}

	tc(r.guest, true)
	tc(r.user1, true)
	tc(r.system, false)
	tc(r.admin, false)
}
