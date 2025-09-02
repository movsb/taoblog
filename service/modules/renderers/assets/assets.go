package assets

import (
	"context"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// 来源：cmd/client/post.go
// 从文章的源代码里面提取出附件列表。
// 参考：docs/usage/文章编辑::自动附件上传

type AssetsParser struct {
	prefix string
	paths  *[]string
}

func New(paths *[]string, prefix string) *AssetsParser {
	return &AssetsParser{
		prefix: prefix,
		paths:  paths,
	}
}

type ctxObj struct{}

// 暂时不支持重复使用，如果需要多个，换 key 为 int 后递增。
func With(ctx context.Context, paths *[]string, prefix string) context.Context {
	if paths == nil {
		panic(`invalid argument`)
	}
	if FromContext(ctx) != nil {
		panic(`dup assets`)
	}
	return context.WithValue(ctx, ctxObj{}, New(paths, prefix))
}

func FromContext(ctx context.Context) *AssetsParser {
	val, _ := ctx.Value(ctxObj{}).(*AssetsParser)
	return val
}

func (p *AssetsParser) TransformHtml(doc *goquery.Document) error {
	if p.paths == nil {
		return nil
	}

	doc.Find(`a,img,iframe,source,audio,video,object`).Each(func(_ int, s *goquery.Selection) {
		if s.HasClass(`static`) {
			return
		}

		var attrName string
		switch s.Nodes[0].Data {
		case `a`:
			attrName = `href`
		case `img`, `iframe`, `source`, `audio`, `video`:
			attrName = `src`
		case `object`:
			attrName = `data`
		}
		attrValue := s.AttrOr(attrName, ``)
		if attrValue == `` {
			return
		}

		if path := parseURL(attrValue, p.prefix); path != `` {
			*p.paths = append(*p.paths, path)
		}
	})

	return nil
}

func parseURL(value string, prefix string) string {
	u, err := url.Parse(value)
	if err != nil {
		return ``
	}

	if u.Scheme != `` || u.Host != `` || u.Path == `` {
		return ``
	}

	if u.Path[0] == '/' {
		if prefix != `` && strings.HasPrefix(u.Path, prefix) {
			return strings.TrimPrefix(u.Path, prefix)
		}
		return ``
	}

	return u.Path
}
