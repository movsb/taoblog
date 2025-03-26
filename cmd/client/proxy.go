package client

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
)

func proxy(ctx context.Context, listen string, home, token string) {
	const ua = `taoblog/proxy`
	parsedHome := utils.Must1(url.Parse(home))
	username, password, _ := strings.Cut(token, `:`)
	req := utils.Must1(http.NewRequestWithContext(ctx, http.MethodPost,
		parsedHome.JoinPath(`/admin/login/basic`).String(),
		bytes.NewReader(utils.Must1(json.Marshal(map[string]string{
			`username`: username,
			`password`: password,
		}))),
	))
	req.Header.Set(`User-Agent`, ua)
	rsp := utils.Must1(http.DefaultClient.Do(req))
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		log.Fatalln(rsp.Status)
	}

	cookies := rsp.Cookies()
	proxy := httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(parsedHome)
			pr.Out.Host = parsedHome.Host
			pr.Out.Header.Set(`User-Agent`, ua)
			for _, c := range cookies {
				pr.Out.AddCookie(c)
			}
		},
	}
	if err := http.ListenAndServe(listen, &proxy); err != http.ErrServerClosed {
		log.Fatalln(err)
	}
}
