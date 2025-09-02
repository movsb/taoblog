package assets

import (
	"context"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yuin/goldmark/ast"
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

// 有些文件会被渲染，光从 html 结果不一定找得到，但是它们确实被引用过。
func (p *AssetsParser) WalkEntering(n ast.Node) (ast.WalkStatus, error) {
	if p.paths == nil {
		return ast.WalkContinue, nil
	}

	if n.Kind() != ast.KindImage {
		return ast.WalkContinue, nil
	}

	img := n.(*ast.Image)
	src := string(img.Destination)

	// markdown 里面只可能是本文的相对路径，不会出现文章编号前缀，
	// 所以这里不需要关心前缀。
	if u := parseURL(src, ""); u != `` {
		*p.paths = append(*p.paths, u)
	}

	return ast.WalkContinue, nil
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
