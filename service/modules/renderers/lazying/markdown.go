package lazying

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

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
func New() any {
	return &_LazyLoadingFrames{}
}

type _LazyLoadingFrames struct{}

func (m *_LazyLoadingFrames) TransformHtml(doc *goquery.Document) error {
	doc.Find(`iframe`).Each(func(i int, s *goquery.Selection) {
		s.Nodes[0].Attr = append(s.Nodes[0].Attr, html.Attribute{Key: `loading`, Val: `lazy`})
	})
	return nil
}
