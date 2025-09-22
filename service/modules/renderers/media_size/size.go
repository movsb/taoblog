package media_size

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	urlpkg "net/url"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
)

type MediaSize struct {
	web gold_utils.WebFileSystem

	localOnly bool
	filter    gold_utils.NodeFilter
}

type Option func(*MediaSize)

func WithLocalOnly() Option {
	return func(ms *MediaSize) {
		ms.localOnly = true
	}
}

func WithNodeFilter(f gold_utils.NodeFilter) Option {
	return func(ms *MediaSize) {
		ms.filter = f
	}
}

// localOnly: 只处理本地图片，不处理网络图片。
// NOTE: 本地文件直接用相对路径指定，不要用 file://。
func New(web gold_utils.WebFileSystem, options ...Option) *MediaSize {
	ms := &MediaSize{
		web:    web,
		filter: func(node *goquery.Selection) bool { return true },
	}
	for _, opt := range options {
		opt(ms)
	}
	return ms
}

func (ms *MediaSize) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img,video`).FilterFunction(func(i int, s *goquery.Selection) bool {
		return ms.filter(s)
	}).Each(func(i int, s *goquery.Selection) {
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
		if q.Has(`cover`) {
			cover := `400px`
			n := utils.DropLast1(strconv.Atoi(q.Get(`cover`)))
			if n > 0 {
				cover = fmt.Sprintf(`%dpx`, n)
			}

			gold_utils.AddStyle(s, fmt.Sprintf(`object-fit: cover; aspect-ratio: 1; width: %s`, cover))

			// 碎碎念比较窄，可以默认 100% 显示。
			// TODO: 移动到 renderers/image 里面处理。
			s.AddClass(`cover`)

			q.Del(`cover`)
		}

		// 可以手动指定宽度。
		prefWidth := 0
		if n, err := strconv.Atoi(q.Get(`w`)); err == nil && n > 0 {
			prefWidth = n
			q.Del(`w`)
		}

		parsedURL.RawQuery = q.Encode()
		s.SetAttr(`src`, parsedURL.String())

		md, err := size(ms.web, parsedURL, ms.localOnly)
		if err != nil {
			if !errors.Is(err, gold_utils.ErrCrossOrigin) && !ms.localOnly {
				log.Println(err, utils.IIF(parsedURL.Scheme == `data`, `(data: url)`, url))
			}
			return
		}

		w, h := md.Width, md.Height

		// 如果指定了 scale，则按比例缩放。
		w, h = int(float64(w)*scale), int(float64(h)*scale)

		// 如果指定了宽度，则按比例缩放高度。
		if prefWidth > 0 {
			// w/h == pw / ph -> ph = h * pw / w
			h = int(float64(h) * float64(prefWidth) / float64(w))
			w = prefWidth
		}

		s.SetAttr(`width`, fmt.Sprint(w))
		s.SetAttr(`height`, fmt.Sprint(h))

		// aspect-ratio 对图片无效。
	})
	doc.Find(`svg`).Each(func(i int, s *goquery.Selection) {
		buf := bytes.NewBuffer(nil)
		// TODO 这里效率可能有点低，直接检测根元素即可。
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

		// aspect-ratio 对 svg 无效。
		// 它是响应式的。
	})
	doc.Find(`iframe`).Each(func(i int, s *goquery.Selection) {
		// 目前只能处理这种大小：<iframe width="560" height="315" ...
		width := utils.DropLast1(strconv.Atoi(s.AttrOr(`width`, `0`)))
		height := utils.DropLast1(strconv.Atoi(s.AttrOr(`height`, `0`)))
		if width <= 0 || height <= 0 {
			return
		}
		gold_utils.AddStyle(s, ratio(width, height))
	})
	return nil
}

func ratio(width, height int) string {
	return fmt.Sprintf(`aspect-ratio: %.6f;`, float32(width)/float32(height))
}

// root: 如果 url 是相对路径，用于指定根文件系统。
func size(fs gold_utils.WebFileSystem, parsedURL *urlpkg.URL, localOnly bool) (*Metadata, error) {
	if (parsedURL.Scheme != "" || parsedURL.Host != "") && localOnly {
		return nil, gold_utils.ErrCrossOrigin
	}

	var r io.Reader

	if parsedURL.Scheme == "" && parsedURL.Host == "" {
		f, err := fs.OpenURL(parsedURL.String())
		if err != nil {
			return nil, err
		}
		defer f.Close()

		if stat, err := f.Stat(); err == nil {
			if g, ok := stat.Sys().(utils.ImageDimensionGetter); ok {
				w, h := g.GetImageDimension()
				if w > 0 && h > 0 {
					return &Metadata{
						Width:  w,
						Height: h,
					}, nil
				}
			}
		}

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

	md, err := All(r)
	if err != nil {
		return nil, err
	}

	return md, nil
}

func All(r io.Reader) (*Metadata, error) {
	dup := utils.MemDupReader(r)

	var errs []error

	for _, d := range []func(r io.Reader) (*Metadata, error){
		normal,
		svg,
	} {
		md, err := d(dup())
		if err == nil {
			return md, nil
		}
		errs = append(errs, err)
	}
	return nil, fmt.Errorf(`no decoder applicable: %w`, errors.Join(errs...))
}
