package comment_notify

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// Chanify notification

type _ChanifyBody struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

// Chanify ...
func Chanify(endpoint, token string, subject string, body string) {
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
		Title: subject,
		Text:  body,
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
	str, _ := ioutil.ReadAll(io.LimitReader(rsp.Body, 1<<10))
	if rsp.StatusCode != 200 {
		log.Printf(`chanify error: %v, %s`, rsp.StatusCode, string(str))
		return
	}
	log.Printf("chanify success: %v", string(str))
}
