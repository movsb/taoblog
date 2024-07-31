package media_size

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	urlpkg "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	gold_utils "github.com/movsb/taoblog/service/modules/renderers/goldutils"
)

type MediaSize struct {
	web gold_utils.WebFileSystem

	localOnly bool
	sizeLimit int
}

type Option func(*MediaSize)

func WithLocalOnly() Option {
	return func(ms *MediaSize) {
		ms.localOnly = true
	}
}

// 图片/视频/大小限制器。
func WithDimensionLimiter(size int) Option {
	return func(ms *MediaSize) {
		ms.sizeLimit = size
	}
}

// localOnly: 只处理本地图片，不处理网络图片。
// NOTE: 本地文件直接用相对路径指定，不要用 file://。
func New(web gold_utils.WebFileSystem, options ...Option) *MediaSize {
	ms := &MediaSize{
		web: web,
	}
	for _, opt := range options {
		opt(ms)
	}
	return ms
}

func (ms *MediaSize) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img`).Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr(`src`, ``)
		if url == "" {
			return
		}

		parsedURL, err := urlpkg.Parse(url)
		if err != nil {
			log.Println(err)
			return
		}

		// TODO 是不是不应该放这里？
		q := parsedURL.Query()
		scale := 1.0
		if n, err := strconv.ParseFloat(q.Get(`scale`), 64); err == nil && n > 0 {
			scale = n
			q.Del(`scale`)
		}
		if n, err := strconv.ParseFloat(q.Get(`s`), 64); err == nil && n > 0 {
			scale = n
			q.Del(`s`)
		}
		parsedURL.RawQuery = q.Encode()
		s.SetAttr(`src`, parsedURL.String())

		md, err := size(ms.web, parsedURL, ms.localOnly)
		if err != nil {
			// TODO 忽略 emoji
			log.Println(err, utils.IIF(parsedURL.Scheme == `data`, `(data: url)`, url))
			return
		}

		w, h := md.Width, md.Height
		w, h = int(float64(w)*scale), int(float64(h)*scale)

		// 暴力且不准确地检测是否文件名中带缩放（早期的实现）。
		// TODO 移除通过 @2x 类似的图片缩放支持。
		for i := 1; i <= 10; i++ {
			scaleFmt := fmt.Sprintf(`@%dx.`, i)
			if strings.Contains(url, scaleFmt) {
				w /= i
				h /= i
				break
			}
		}

		s.SetAttr(`width`, fmt.Sprint(w))
		s.SetAttr(`height`, fmt.Sprint(h))

		limit(s, ms.sizeLimit, w, h)
	})
	doc.Find(`svg`).Each(func(i int, s *goquery.Selection) {
		buf := bytes.NewBuffer(nil)
		goquery.Render(buf, s)
		md, err := svg(buf)
		if err != nil {
			log.Println(err)
			return
		}
		if _, ok := s.Attr(`width`); !ok {
			s.SetAttr(`width`, fmt.Sprintf(`%d`, md.Width))
		}
		if _, ok := s.Attr(`height`); !ok {
			s.SetAttr(`height`, fmt.Sprintf(`%d`, md.Height))
		}
		var width, height int
		fmt.Sscanf(s.AttrOr(`width`, `0`), "%d", &width)
		fmt.Sscanf(s.AttrOr(`height`, `0`), "%d", &height)
		limit(s, ms.sizeLimit, width, height)
	})
	doc.Find(`iframe`).Each(func(i int, s *goquery.Selection) {
		// 目前只能处理这种大小：<iframe width="560" height="315" ...
		width := utils.DropLast1(strconv.Atoi(s.AttrOr(`width`, `0`)))
		height := utils.DropLast1(strconv.Atoi(s.AttrOr(`height`, `0`)))
		if width <= 0 || height <= 0 {
			return
		}
		style := s.AttrOr(`style`, ``)
		style += `aspect-ratio:` + ratio(width, height)
		s.SetAttr(`style`, style)
		limit(s, ms.sizeLimit, width, height)
	})
	return nil
}

func ratio(width, height int) string {
	// 欧几里德法求最大公约数
	gcd := func(a, b int) int {
		for b != 0 {
			a, b = b, a%b
		}
		return a
	}(width, height)
	return fmt.Sprintf(`%d/%d;`, width/gcd, height/gcd)
}

func limit(s *goquery.Selection, limit, width, height int) {
	if limit > 0 {
		// == 的情况也一起处理了。
		if width >= height {
			s.AddClass(`landscape`)
			if width > limit {
				s.AddClass(`too-wide`)
			}
		} else {
			s.AddClass(`portrait`)
			if height > limit {
				s.AddClass(`too-high`)
			}
		}
	}
}

// root: 如果 url 是相对路径，用于指定根文件系统。
func size(fs gold_utils.WebFileSystem, parsedURL *url.URL, localOnly bool) (*Metadata, error) {
	if (parsedURL.Scheme != "" || parsedURL.Host != "") && localOnly {
		return nil, errors.New(`not for network images`)
	}

	var r io.Reader

	if parsedURL.Scheme == "" && parsedURL.Host == "" {
		f, err := fs.OpenURL(parsedURL.String())
		if err != nil {
			return nil, err
		}
		defer f.Close()
		r = f
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
		if err != nil {
			return nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		r = resp.Body
	}

	md, err := all(r)
	if err != nil {
		return nil, err
	}

	return md, nil
}

func all(r io.Reader) (*Metadata, error) {
	dup := utils.MemDupReader(r)

	var errs []error

	for _, d := range []func(r io.Reader) (*Metadata, error){
		normal,
		svg,
		avif,
	} {
		md, err := d(dup())
		if err == nil {
			return md, nil
		}
		errs = append(errs, err)
	}
	return nil, fmt.Errorf(`no decoder applicable: %w`, errors.Join(errs...))
}
