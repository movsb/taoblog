package footnotes

import (
	"embed"
	"fmt"
	"hash/fnv"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

//go:generate sass --style compressed --no-source-map style.scss style.css

//go:embed style.css
var _embed embed.FS
var _root = utils.NewOSDirFS(dir.SourceAbsoluteDir().Join())

func init() {
	dynamic.RegisterInit(func() {
		const module = `footnotes`
		dynamic.WithRoots(module, nil, nil, _embed, _root)
		dynamic.WithStyles(module, `style.css`)
	})
}

// prefix: 用于区分不同文章、不同评论的前缀。
//
// 官方默认形式如： fn:1
//
// 如果不加区分，不同文章、不同评论将完全共用ID空间，出现混乱。
//
// 方案：`${type}:${post_id}-`，其中：type = a | c，分别表示文章和评论。
//
// 如： `a:100-fn:1`
//
// 但是这样有点丑，改用直接统一哈希的方式。
//

type Type int

const (
	Article Type = iota + 1
	Comment
)

type Extender struct {
	fn goldmark.Extender
}

// 因为文章和评论编号可能相同，所以用 Type 区分名字空间。
func New(typ Type, id int) goldmark.Extender {
	return &Extender{
		fn: extension.NewFootnote(
			extension.WithFootnoteBacklinkHTML(`^`),
			extension.WithFootnoteIDPrefix(fmt.Sprintf(`type:%d:%d-`, typ, id)),
		),
	}
}

func (e *Extender) Extend(m goldmark.Markdown) {
	e.fn.Extend(m)
}

type Hash struct {
	m map[uint32]string
}

func NewHash() *Hash {
	return &Hash{m: map[uint32]string{}}
}

type ID uint32

func (h *Hash) Hash(id string) ID {
	alg := fnv.New32a()
	alg.Write([]byte(id))
	sum := alg.Sum32()
	for {
		if val, ok := h.m[sum]; ok {
			if val == id {
				// 有且值相等，直接使用
				return ID(sum)
			} else {
				// 不相等，说明哈希冲突，
				// 简单开放寻址到下一个。
				sum++
				continue
			}
		} else {
			// 未被使用，直接存。
			h.m[sum] = id
			return ID(sum)
		}
	}
}

func (id ID) String() string {
	return fmt.Sprintf(`fn:%08x`, uint32(id))
}

func (e *Extender) TransformHtml(doc *goquery.Document) error {
	hash := NewHash()

	forID := func(i int, s *goquery.Selection) {
		id := s.AttrOr(`id`, ``)
		new := hash.Hash(id)
		s.SetAttr(`id`, new.String())
	}
	forHref := func(i int, s *goquery.Selection) {
		id := s.AttrOr(`href`, `#`)[1:]
		new := hash.Hash(id)
		s.SetAttr(`href`, `#`+new.String())
	}

	doc.Find(`div.footnotes > ol > li`).Each(func(i int, s *goquery.Selection) {
		forID(i, s)
		s.Find(`a.footnote-backref`).Each(forHref)
	})
	doc.Find(`a.footnote-ref`).Each(func(i int, s *goquery.Selection) {
		forHref(i, s)
		s.Closest(`sup`).Each(forID)
	})

	return nil
}
