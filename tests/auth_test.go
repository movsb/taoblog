package e2e_test

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/micros/auth/cookies"
)

func TestRefreshCookies(t *testing.T) {
	r := Serve(t.Context())

	cookies.SetMaxAge(time.Second * 2)

	loginURL := r.server.JoinPath(`admin/login/basic`)
	rsp := utils.Must1(http.Post(
		loginURL, `application/json`,
		strings.NewReader(fmt.Sprintf(`
{
	"username": "%d",
	"password": %s
}
	`,
			r.user1ID,
			string(utils.Must1(json.Marshal(r.user1Password))),
		))))

	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		t.Fatalf(`登录失败。`)
	}

	getCookie := func(rsp *http.Response) string {
		value := utils.Map(utils.Filter(rsp.Cookies(), func(c *http.Cookie) bool {
			return c.Name == cookies.CookieNameLogin
		}), func(c *http.Cookie) string {
			return c.Value
		})
		if len(value) > 0 {
			return value[0]
		}
		return ``
	}

	reqCookie := func(provide string) string {
		req := utils.Must1(http.NewRequest(`GET`, r.server.JoinPath(), nil))
		if provide != `` {
			req.AddCookie(&http.Cookie{
				Name:  cookies.CookieNameLogin,
				Value: provide,
			})
		}
		rsp := utils.Must1(http.DefaultClient.Do(req))
		rsp.Header.Write(log.Writer())
		return getCookie(rsp)
	}

	c1 := getCookie(rsp)
	if c1 == `` {
		t.Fatal(`no value`)
	}

	time.Sleep(time.Second)
	c2 := reqCookie(c1)
	if c2 == c1 || c2 == `` {
		t.Fatal(`bad cookie refresh`)
	}
}
