package comment_notify

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Chanify notification

type _ChanifyBody struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

var _chanifyTemplate = template.Must(template.New(`chanify`).Parse(`
{{- .Content }}

文章链接：{{ .Link }}
评论作者：{{ .Author }}
评论日期：{{ .Date }}
作者邮箱：{{ .Email }}
作者主页：{{ .HomePage }}
`))

func executeChanifyTemplate(data *AdminData) string {
	b := bytes.NewBuffer(nil)
	if err := _chanifyTemplate.Execute(b, data); err != nil {
		log.Println(err)
		return ""
	}
	return b.String()
}

// Chanify ...
func Chanify(endpoint, token string, data *AdminData) {
	if p := strings.Index(endpoint, "://"); p == -1 {
		endpoint = `http://` + endpoint
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		log.Printf(`chanify error: %v`, err)
		return
	}
	u.Path = filepath.Join(u.Path, `/v1/sender/`, token)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	cb := _ChanifyBody{
		Title: data.Title,
		Text:  executeChanifyTemplate(data),
	}
	b, _ := json.Marshal(cb)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(b))
	if err != nil {
		log.Printf(`new request failed: %v`, err)
		return
	}
	req.Header.Set(`Content-Type`, "application/json")
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf(`request error: %v`, err)
		return
	}
	defer rsp.Body.Close()
	str, _ := io.ReadAll(io.LimitReader(rsp.Body, 1<<10))
	if rsp.StatusCode != 200 {
		log.Printf(`chanify error: %v, %s`, rsp.StatusCode, string(str))
		return
	}
	log.Printf("chanify success: %v", string(str))
}
