package client

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/movsb/taoblog/modules/auth/cookies"
	"github.com/movsb/taoblog/modules/utils"
)

func proxy(ctx context.Context, listen string, home, token string) {
	parsedHome := utils.Must1(url.Parse(home))
	proxy := httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(parsedHome)
			pr.Out.Host = parsedHome.Host
			pr.Out.Header.Del(`Cookie`)
			pr.Out.Header.Add(`Authorization`, fmt.Sprintf(`%s %s`, cookies.TokenName, token))
		},
	}
	log.Println(`HTTP:`, fmt.Sprintf(`http://%s`, listen))
	if err := http.ListenAndServe(listen, &proxy); err != http.ErrServerClosed {
		log.Fatalln(err)
	}
}
