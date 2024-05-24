package renderers

import (
	"bytes"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yuin/goldmark/ast"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// 后面统一改成 Option。
type Option2 = any

type EnteringWalker interface {
	WalkEntering(n ast.Node) (ast.WalkStatus, error)
}

// 非常低效的接口。。。
type HtmlFilter interface {
	FilterHtml(doc *html.Node) ([]byte, error)
}
type HtmlPrettifier interface {
	PrettifyHtml(doc *html.Node) ([]byte, error)
}

type HtmlTransformer interface {
	TransformHtml(doc *goquery.Document) error
}

// -----------------------------------------------------------------------------

func appendClass(node ast.Node, name string) {
	any, found := node.AttributeString(name)
	if list, ok := any.(string); !found || ok {
		if list == "" {
			list = name
		} else {
			list += " "
			list += name
		}
		node.SetAttributeString(`class`, list)
	}
}

func Testing() Option {
	return func(me *_Markdown) error {
		me.testing = true
		return nil
	}
}

// -----------------------------------------------------------------------------

// 获取从 Markdown 中解析得到的一级标题。
func WithTitle(title *string) OptionNoError {
	return func(me *_Markdown) {
		me.title = title
	}
}

// -----------------------------------------------------------------------------

type _ReserveListItemMarkerStyle struct{}

var knownListItemMarkers = map[byte]string{
	'-': `minus`,
	'+': `plus`,
	'*': `asterisk`,
	'.': `period`,
	')': `parenthesis`,
}

func (*_ReserveListItemMarkerStyle) WalkEntering(n ast.Node) (ast.WalkStatus, error) {
	switch typed := n.(type) {
	case *ast.List:
		if class, ok := knownListItemMarkers[typed.Marker]; ok {
			appendClass(typed, `marker-`+class)
		}
	}
	return ast.WalkContinue, nil
}

// 保留列表样式。
//
// 只是增加类名，前端通过类名自行决定怎么展示。
func WithReserveListItemMarkerStyle() Option2 {
	return &_ReserveListItemMarkerStyle{}
}

// -----------------------------------------------------------------------------

var prettifierStrings = map[string]string{
	`img`:    `图片`,
	`table`:  `表格`,
	`video`:  `视频`,
	`audio`:  `音频`,
	`canvas`: `画布`,
	`embed`:  `对象`,
	`map`:    `地图`,
	`object`: `对象`,
	`script`: `脚本`,
	`svg`:    `图片`,
	`code`:   `代码`,
}

var prettifierFuncs = map[string]func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus{
	`a`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		hrefIndex := slices.IndexFunc(node.Attr, func(attr html.Attribute) bool { return attr.Key == `href` })
		if hrefIndex >= 0 && node.FirstChild != nil {
			var label string
			if node.FirstChild.Type == html.TextNode {
				label = node.FirstChild.Data
			} else if node.FirstChild.DataAtom == atom.Code {
				label = node.FirstChild.FirstChild.Data
			}
			// TODO 需要解 URL 码吗？
			if label != node.Attr[hrefIndex].Val {
				buf.WriteString(label)
				return ast.WalkSkipChildren
			}
		}
		buf.WriteString(`[链接]`)
		return ast.WalkSkipChildren
	},
	`div`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		for _, a := range node.Attr {
			if strings.ToLower(a.Key) == `class` {
				if strings.Contains(a.Val, `footnotes`) {
					return ast.WalkSkipChildren
				}
			}
		}
		return ast.WalkContinue
	},
	`pre`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		if node.FirstChild != nil && node.FirstChild.NextSibling == nil && node.FirstChild.Data == `code` {
			buf.WriteString(`[代码]`)
			return ast.WalkSkipChildren
		}
		return ast.WalkContinue
	},
	`iframe`: func(buf *bytes.Buffer, node *html.Node) ast.WalkStatus {
		for _, a := range node.Attr {
			if strings.ToLower(a.Key) == `src` {
				if u, err := url.Parse(a.Val); err == nil {
					switch strings.ToLower(u.Hostname()) {
					case `www.youtube.com`:
						buf.WriteString(`[油管]`)
						return ast.WalkSkipChildren
					}
				}
			}
		}
		buf.WriteString(`[页面]`)
		return ast.WalkSkipChildren
	},
}

type _ContentPrettifier struct{}

func (*_ContentPrettifier) PrettifyHtml(doc *html.Node) ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	var walk func(buf *bytes.Buffer, node *html.Node)
	walk = func(buf *bytes.Buffer, node *html.Node) {
		walkStatus := ast.WalkContinue
		switch node.Type {
		case html.TextNode:
			buf.WriteString(node.Data)
		case html.ElementNode:
			walkStatus = ast.WalkContinue
			if f, ok := prettifierStrings[node.Data]; ok {
				// 简短的代码直接显示。
				if node.Data == `code` &&
					node.FirstChild != nil &&
					node.FirstChild.Type == html.TextNode &&
					node.FirstChild.NextSibling == nil &&
					len(node.FirstChild.Data) <= 16 {
					buf.WriteString(node.FirstChild.Data)
					walkStatus = ast.WalkSkipChildren
				} else {
					buf.WriteString(fmt.Sprintf(`[%s]`, f))
					walkStatus = ast.WalkSkipChildren
				}
			} else if f, ok := prettifierFuncs[node.Data]; ok {
				walkStatus = f(buf, node)
			}
		}
		if walkStatus == ast.WalkContinue {
			for c := node.FirstChild; c != nil; c = c.NextSibling {
				walk(buf, c)
			}
		}
	}

	walk(buf, doc)

	return buf.Bytes(), nil
}

// 简化并美化 Markdown 的展示。
// 比如：
// - 不显示复杂的链接、图片、表格、代码块等元素。
// - 不显示脚注。
//
// NOTE 因为 Markdown 只能解析而不能还原，所以这里处理的是 HTML 内容。
// HTML 可以在 NODE 之间等价转换。
// 而由于 Markdown 目前在转换成 HTML 时采用了后端渲染代码。
// 所以 chroma 把 code 包裹在了 table 中。需要特别处理。
//
// https://yari-demos.prod.mdn.mozit.cloud/en-US/docs/Web/HTML/Inline_elements
func WithHtmlPrettifier() Option2 {
	return &_ContentPrettifier{}
}

// -----------------------------------------------------------------------------

// 油管的分享视频 iframe 竟然默认不是 lazy lading 的，有点儿无语😓。
// 目前碎碎念是全部加载的，有好几个视频，会严重影响页面加载速度。
//
// 做法是解析 HTML Block，判断是否为 iframe，然后添加属性。
//
// NOTE：Markdown 虽然允许 html 和  markdown 交叉混写。但是处理这种交叉的内容
// 非常复杂（涉及不完整 html 的解析与还原），所以暂时不支持这种情况。
// 这种情况很少，像是 <iframe 油管视频> 都是在一行内。就算可以多行，也不会和 markdown 交织。
// 虽然 iframe 是 inline 类型的元素，但是应该没人放在段落内吧？都是直接粘贴成为一段的。否则不能处理。
//
// https://developer.mozilla.org/en-US/docs/Web/Performance/Lazy_loading#loading_attribute
func WithLazyLoadingFrames() Option2 {
	return &_LazyLoadingFrames{}
}

type _LazyLoadingFrames struct{}

func (m *_LazyLoadingFrames) FilterHtml(doc *html.Node) ([]byte, error) {

	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		switch node.Type {
		case html.ElementNode:
			if node.Data == `iframe` {
				node.Attr = append(node.Attr, html.Attribute{Key: `loading`, Val: `lazy`})
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)

	return renderHtmlDoc(doc)
}

// -----------------------------------------------------------------------------

// 对 WithUseAbsolutePaths 的补充。
// 其实含义相同，只是换了个更正确的名字。
// 上述只能针对 md 的 img 和 a，没法针对用 html
// 插入的 audio / video / iframe / object。
func WithRootedPaths(base string) Option2 {
	return &_RootedPaths{
		root: base,
	}
}

type _RootedPaths struct {
	root string
}

func (m *_RootedPaths) FilterHtml(doc *html.Node) ([]byte, error) {
	find := func(node *html.Node, name string) int {
		for i, a := range node.Attr {
			if a.Key == name {
				return i
			}
		}
		return -1
	}
	modify := func(val *string) bool {
		if u, err := url.Parse(*val); err == nil {
			if u.Scheme == "" && u.Host == "" && !filepath.IsAbs(u.Path) {
				u.Path = path.Join(m.root, u.Path)
				*val = u.EscapedPath()
				return true
			}
		}
		return false
	}

	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		switch node.Type {
		case html.ElementNode:
			name := ""
			switch node.Data {
			case `iframe`, `source`, `audio`, `video`:
				name = `src`
			case `object`:
				name = `data`
			}
			if name != "" {
				if index := find(node, name); index >= 0 {
					// old := node.Attr[index].Val
					if modify(&node.Attr[index].Val) {
						// // 修改后保留一份原始路径供其它地方使用。
						// node.Attr = append(node.Attr, html.Attribute{
						// 	Key: `data-path`,
						// 	Val: old,
						// })
					}
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)

	return renderHtmlDoc(doc)
}

// -----------------------------------------------------------------------------

// 图片/视频/大小限制器。
func WithMediaDimensionLimiter(max int) Option2 {
	return &_MediaDimensionLimiter{max: max}
}

type _MediaDimensionLimiter struct {
	max int
}

func (m *_MediaDimensionLimiter) FilterHtml(doc *html.Node) ([]byte, error) {
	find := func(node *html.Node, name string) int {
		for i, a := range node.Attr {
			if a.Key == name {
				return i
			}
		}
		return -1
	}
	add := func(node *html.Node, name string) {
		if p := find(node, `class`); p >= 0 {
			if strings.Contains(node.Attr[p].Val, name) {
				panic(`已经包含了`)
			}
			node.Attr[p].Val += ` ` + name
		} else {
			node.Attr = append(node.Attr, html.Attribute{
				Key: `class`,
				Val: name,
			})
		}
	}
	modify := func(node *html.Node, width, height *string) {
		var w, h int
		fmt.Sscanf(*width, "%d", &w)
		fmt.Sscanf(*height, "%d", &h)
		if w > h {
			add(node, `landscape`)
			if w > m.max {
				add(node, `too-wide`)
			}
		} else if h > w {
			add(node, `portrait`)
			if h > m.max {
				add(node, `too-high`)
			}
		}
	}

	var walk func(node *html.Node)
	walk = func(node *html.Node) {
		switch node.Type {
		case html.ElementNode:
			switch node.Data {
			case `video`, `img`:
				indexWidth := find(node, `width`)
				indexHeight := find(node, `height`)
				if indexWidth >= 0 && indexHeight >= 0 {
					modify(node, &node.Attr[indexWidth].Val, &node.Attr[indexHeight].Val)
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(doc)

	return renderHtmlDoc(doc)
}
