package colors

import (
	"embed"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

//go:generate go run ./gen/gen.go colors.yaml colors.scss

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

func init() {
	dynamic.RegisterInit(func() {
		const module = `colors`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

type _Colors struct{}

func New() *_Colors {
	return &_Colors{}
}

func (m *_Colors) TransformHtml(doc *goquery.Document) error {
	doc.Find(`color`).Each(func(i int, s *goquery.Selection) {
		transform(s)
	})
	return nil
}

var (
	reColor = regexp.MustCompile(`^([[:alpha:]]+)?(?::([[:alpha:]]+))?$`)
)

func transform(node *goquery.Selection) {
	hn := node.Nodes[0]
	attrsToRemove := []string{}
	for _, attr := range hn.Attr {
		switch key := strings.ToLower(attr.Key); key {
		case `fg`, `bg`:
			attrsToRemove = append(attrsToRemove, attr.Key)

			// 可能无效或尝试注入的值。
			if strings.ContainsAny(attr.Val, `;'"\`) {
				break
			}

			styles := node.AttrOr(`style`, ``)

			if key == `fg` {
				styles += fmt.Sprintf(`color:%s;`, attr.Val)
			} else {
				styles += fmt.Sprintf(`background-color:%s;`, attr.Val)
			}

			node.SetAttr(`style`, styles)
		default:
			matches := reColor.FindStringSubmatch(attr.Key)
			if matches == nil {
				break
			}

			attrsToRemove = append(attrsToRemove, attr.Key)

			if attr.Val != `` {
				break
			}

			node.AddClass(`color`)
			if fg := matches[1]; fg != `` && hasColor(fg) {
				node.AddClass(`fg-` + fg)
			}
			if bg := matches[2]; bg != `` && hasColor(bg) {
				node.AddClass(`bg-` + bg)
			}
		}
	}

	for _, a := range attrsToRemove {
		node.RemoveAttr(a)
	}

	// goquery 有个bug，设置的类名有多出的空格字符。
	if c := node.AttrOr(`class`, ``); c != `` {
		node.SetAttr(`class`, strings.Join(strings.Fields(c), ` `))
	}

	if shouldBeBlockElement(node) {
		hn.DataAtom = atom.Div
		hn.Data = `div`
	} else {
		hn.DataAtom = atom.Span
		hn.Data = `span`
	}
}

// 判断当前元素是否更应该被当作块级元素。
//
// 判断规则是 ad-hoc 的。
func shouldBeBlockElement(s *goquery.Selection) bool {
	// 如果没有父元素
	if s.Parent().Length() <= 0 {
		return true
	}

	// fast path：如果父亲是 <p>。
	switch s.Parent().Nodes[0].DataAtom {
	case atom.P, atom.Li, atom.Td, atom.Th:
		return false
	}

	// 近亲有文本内容，大概率是内联元素。
	for _, child := range s.Parent().Contents().Nodes {
		if child.Type == html.TextNode && strings.TrimSpace(child.Data) != `` {
			return false
		}
	}

	// 如果内部有常用块级元素。
	if s.Find(`div,p,pre,blockquote,ul,ol,li,table,h1,h2,h3,h4,h5,h6`).Length() > 0 {
		return true
	}

	return true
}

//go:embed colors.yaml
var colors []byte

type Color struct {
	Name string `yaml:"name"`
	Hex  string `yaml:"hex"`
}

var oncePalette = sync.OnceValue(func() map[string]string {
	var cs []Color
	rc := map[string]string{}
	utils.Must(yaml.Unmarshal(colors, &cs))
	for _, c := range cs {
		rc[strings.ToLower(c.Name)] = c.Hex
	}
	return rc
})

func hasColor(name string) bool {
	_, has := oncePalette()[strings.ToLower(name)]
	return has
}
