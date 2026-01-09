package gold_utils

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"net/url"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	gold_html "github.com/yuin/goldmark/renderer/html"
)

func AddClass(node ast.Node, classes ...string) {
	var class string
	if any, ok := node.AttributeString(`class`); ok {
		if str, ok := any.(string); ok {
			class = str
		}
	}
	classNames := strings.Fields(class)
	classNames = append(classNames, classes...)
	slices.Sort(classNames)
	classNames = slices.Compact(classNames)

	node.SetAttributeString(`class`, strings.Join(classNames, ` `))
}

// URL 引用文件系统。
//
// 场景：前端相对路径链接。
type WebFileSystem interface {
	// 比如页面地址是：/page/
	// 如果 url 是：1.txt，则打开 /page/1.txt
	// 如果 url 是：/other/2.txt，则打开 /other/2.txt
	OpenURL(url string) (fs.File, error)
	// relative: 结果是否只保留相对路径（前提是同源的情况下）。
	Resolve(url string, relative bool) (*url.URL, error)
}

func NewWebFileSystem(root fs.FS, base *url.URL) WebFileSystem {
	return &_WebFileSystem{
		root: root,
		base: base,
	}
}

type _WebFileSystem struct {
	root fs.FS
	base *url.URL
}

func (fs *_WebFileSystem) Resolve(url_ string, relative bool) (*url.URL, error) {
	u, err := url.Parse(url_)
	if err != nil {
		return nil, err
	}
	ref := fs.base.ResolveReference(u)
	if relative && ref.Host == fs.base.Host {
		ref = &url.URL{
			Path:     ref.Path,
			RawQuery: ref.RawQuery,
		}
	}
	return ref, nil
}

func (fs *_WebFileSystem) OpenURL(url_ string) (fs.File, error) {
	ref, err := fs.Resolve(url_, false)
	if err != nil {
		return nil, err
	}
	// 即使 base 不包含 host 也满足。
	if !strings.EqualFold(fs.base.Host, ref.Host) {
		// if ref.Scheme != `data` {
		// 	log.Println(`fs: url:`, url_)
		// }
		return nil, ErrCrossOrigin
	}
	// fs.FS 不能以 / 开头。
	return fs.root.Open(ref.Path[1:])
}

var ErrCrossOrigin = errors.New(`file is from another origin, cannot be opened`)

type HtmlTransformer interface {
	TransformHtml(doc *goquery.Document) error
}

func ApplyHtmlTransformers(raw []byte, trs ...HtmlTransformer) ([]byte, error) {
	if len(trs) <= 0 {
		return raw, nil
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(raw))
	if err != nil {
		return raw, err
	}

	for _, tr := range trs {
		if err := tr.TransformHtml(doc); err != nil {
			return raw, err
		}
	}

	headChildren := doc.Find(`head`).Children()
	bodyChildren := doc.Find(`body`).Children()

	buf := bytes.NewBuffer(nil)
	var outErr error

	headChildren.EachWithBreak(func(i int, s *goquery.Selection) bool {
		if err := goquery.Render(buf, s); err != nil {
			outErr = err
			return false
		}
		return true
	})
	if outErr != nil {
		return raw, outErr
	}

	bodyChildren.EachWithBreak(func(i int, s *goquery.Selection) bool {
		if err := goquery.Render(buf, s); err != nil {
			outErr = err
			return false
		}
		buf.WriteRune('\n')
		return true
	})
	if outErr != nil {
		return raw, outErr
	}

	return buf.Bytes(), nil
}

func AttrIntOrDefault(attrs parser.Attributes, name string, def int) int {
	if a, ok := attrs.Find([]byte(name)); ok {
		if i, ok := a.(float64); ok {
			return int(i)
		}
	}
	return def
}

func AddStyle(s *goquery.Selection, style string) {
	old := strings.TrimSpace(s.AttrOr(`style`, ``))
	if old != `` && !strings.HasSuffix(old, `;`) {
		old += `;`
	}
	old += style
	s.SetAttr(`style`, old)
}

// 注入一张立即出错的图片，因可以触发 onerror 事件，所以注入可执行的脚本。
func InjectImage(fn string) string {
	// display: 不影响 document flow
	// static: 不处理事件等
	// eager: lazy 的时候过滤不处理
	return `<img style="display:none;" class="static" loading="eager" src="data:image/png;base64,AAAA" onerror="` + fn + `(this)">`
}

func RenderError(w io.Writer, err error) {
	fmt.Fprint(w, `<div class="render-error">`)
	fmt.Fprint(w, html.EscapeString(strings.TrimSpace(err.Error())))
	fmt.Fprint(w, `</div>`)
}

var htmlWriter = gold_html.NewWriter()

// 见 TestTitle
func RenderToText(rawMarkdown []byte) string {
	buf := bytes.NewBuffer(nil)
	bw := bufio.NewWriter(buf)
	htmlWriter.Write(bw, rawMarkdown)
	bw.Flush()
	return html.UnescapeString(buf.String())
}
