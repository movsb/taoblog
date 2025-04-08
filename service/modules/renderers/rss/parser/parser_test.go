package rss_parser_test

import (
	"context"
	_ "embed"
	"encoding/xml"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	rss_parser "github.com/movsb/taoblog/service/modules/renderers/rss/parser"
)

//go:embed testdata/twofei.rss
var twofeiRss []byte

func TestParseSubscription(t *testing.T) {
	var sub rss_parser.RSS
	if err := xml.Unmarshal(twofeiRss, &sub); err != nil {
		t.Fatal(err)
	}
	eq := func(a, b string) {
		if a != b {
			t.Fatal(a, b)
		}
	}
	eq(sub.Version, `2.0`)
	eq(sub.Channel.Title.String(), `陪她去流浪`)
	eq(sub.Channel.Link.String(), `https://blog.twofei.com`)
	eq(sub.Channel.Description.String(), `博客建站🔟周年快乐🎉！圣诞快乐🎄！`)
	eq(sub.Channel.Items[1].Title.String(), `新年快乐，建站十周年快乐🎉`)
	eq(sub.Channel.Items[1].PubDate.Format(time.RFC1123), `Wed, 01 Jan 2025 02:03:00 CST`)
	eq(sub.Channel.Items[1].Description.String(), `<p></p>`)
}

//go:embed testdata/twofei.opml
var twofeiOpml []byte

func TestParseOpml(t *testing.T) {
	var sub rss_parser.OPML
	if err := xml.Unmarshal(twofeiOpml, &sub); err != nil {
		t.Fatal(err)
	}
	eq := func(a, b string) {
		if a != b {
			t.Fatal(a, b)
		}
	}
	eq(sub.Head.Title.String(), `momo`)
	eq(sub.Body.Outlines[0].Title.String(), `FC;NES`)
	eq(sub.Body.Outlines[0].Outlines[0].XmlUrl.String(), `http://cah4e3.wordpress.com/feed/`)
	eq(sub.Body.Outlines[1].Title.String(), `陪她去流浪`)
}

//go:embed testdata/feed.xml
var feedXML []byte

func TestParseFeed(t *testing.T) {
	var sub rss_parser.Feed
	if err := xml.Unmarshal(feedXML, &sub); err != nil {
		t.Fatal(err)
	}
	eq := func(a, b string) {
		if a != b {
			t.Fatal(a, b)
		}
	}
	eq(sub.Title.String(), `阮一峰的网络日志`)
	eq(sub.Updated.Format(time.RFC3339), `2025-04-03T08:02:39Z`)
	eq(sub.Entries[0].Title.String(), `科技爱好者周刊（第 343 期）：如何阻止 AI 爬虫`)
	eq(sub.Entries[0].Link.Href, `http://www.ruanyifeng.com/blog/2025/03/weekly-issue-343.html`)
	eq(sub.Entries[0].Published.Format(time.RFC3339), `2025-03-28T00:09:51Z`)
	eq(sub.Entries[0].Summary.Data.String(), `这里记录每周值得分享的科技内容，周五发布。（[通知] 下周清明假期，周刊休息。）...`)
	eq(sub.Entries[0].Content.Type, `html`)
	eq(strings.TrimSpace(sub.Entries[0].Content.Data.String()), `<p>这里记的科技内容，周五发布。</p>`)
}

// // go:embed testdata/feedly.opml
var feedlyOpml []byte

func TestParseOpml2(t *testing.T) {
	t.SkipNow()
	var opml rss_parser.OPML
	if err := xml.Unmarshal(feedlyOpml, &opml); err != nil {
		t.Fatal(err)
	}
	wg := sync.WaitGroup{}
	opml.Each(func(title, url string) {
		wg.Add(1)
		go func(title, url string) {
			defer wg.Done()
			testURL(t, title, url)
		}(title, url)
	})
	wg.Wait()
}

func testURL(t *testing.T, title, url string) (outAny any) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err, url)
		return
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		t.Error(rsp.Status, url)
		return
	}
	parsed, err := (rss_parser.Parse(rsp.Body))
	if err != nil {
		t.Error(`不能解析为 rss 或 feed。`, title, url, err)
		return
	}
	return parsed
}

func TestParseDate(t *testing.T) {
	tests := []string{
		`2019-03-17T00:00:00+00:00`,
	}
	for _, te := range tests {
		var d rss_parser.Date
		utils.Must(xml.Unmarshal([]byte(`<d>`+te+`</d>`), &d))
	}
}

func TestURL(t *testing.T) {
	t.SkipNow()
	p := testURL(t, `1`, `https://www.worldhello.net/atom.xml`)
	f := p.(*rss_parser.Feed)
	t.Log(f.Entries[0].Published)
}
