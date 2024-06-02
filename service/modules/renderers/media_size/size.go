package media_size

import (
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
	fs        gold_utils.URLReferenceFileSystem
	localOnly bool
}

// localOnly: 只处理本地图片，不处理网络图片。
// NOTE: 本地文件直接用相对路径指定，不要用 file://。
func New(fs gold_utils.URLReferenceFileSystem, localOnly bool) *MediaSize {
	return &MediaSize{
		fs:        fs,
		localOnly: localOnly,
	}
}

func (ms *MediaSize) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img`).Each(func(i int, s *goquery.Selection) {
		url := s.AttrOr(`src`, ``)
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

		md, err := size(ms.fs, parsedURL, ms.localOnly)
		if err != nil {
			log.Println(url, err)
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
	})
	return nil
}

// root: 如果 url 是相对路径，用于指定根文件系统。
func size(fs gold_utils.URLReferenceFileSystem, parsedURL *url.URL, localOnly bool) (*Metadata, error) {
	if (parsedURL.Scheme != "" || parsedURL.Host != "") && localOnly {
		return nil, errors.New(`not for network images`)
	}

	var r io.Reader

	if parsedURL.Scheme == "" && parsedURL.Host == "" {
		f, err := fs.OpenReference(parsedURL.String())
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

	for _, d := range []func(r io.Reader) (*Metadata, error){
		normal,
		svg,
		avif,
	} {
		md, err := d(dup())
		if err == nil {
			return md, nil
		}
		log.Println(err)
	}
	return nil, errors.New(`no decoder applicable`)
}
