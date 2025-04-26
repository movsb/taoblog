package service

import (
	"context"
	"fmt"
	"html"

	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/renderers"
	"github.com/movsb/taoblog/service/modules/renderers/alerts"
	"github.com/movsb/taoblog/service/modules/renderers/custom_break"
	"github.com/movsb/taoblog/service/modules/renderers/echarts"
	"github.com/movsb/taoblog/service/modules/renderers/emojis"
	"github.com/movsb/taoblog/service/modules/renderers/exif"
	"github.com/movsb/taoblog/service/modules/renderers/footnotes"
	"github.com/movsb/taoblog/service/modules/renderers/friends"
	"github.com/movsb/taoblog/service/modules/renderers/gallery"
	"github.com/movsb/taoblog/service/modules/renderers/genealogy"
	"github.com/movsb/taoblog/service/modules/renderers/gold_utils"
	"github.com/movsb/taoblog/service/modules/renderers/graph_viz"
	"github.com/movsb/taoblog/service/modules/renderers/hashtags"
	"github.com/movsb/taoblog/service/modules/renderers/highlight"
	"github.com/movsb/taoblog/service/modules/renderers/iframe"
	"github.com/movsb/taoblog/service/modules/renderers/image"
	"github.com/movsb/taoblog/service/modules/renderers/invalid_scheme"
	"github.com/movsb/taoblog/service/modules/renderers/link_target"
	"github.com/movsb/taoblog/service/modules/renderers/list_markers"
	"github.com/movsb/taoblog/service/modules/renderers/live_photo"
	"github.com/movsb/taoblog/service/modules/renderers/math"
	"github.com/movsb/taoblog/service/modules/renderers/media_size"
	"github.com/movsb/taoblog/service/modules/renderers/media_tags"
	"github.com/movsb/taoblog/service/modules/renderers/page_link"
	"github.com/movsb/taoblog/service/modules/renderers/pikchr"
	"github.com/movsb/taoblog/service/modules/renderers/plantuml"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
	"github.com/movsb/taoblog/service/modules/renderers/rooted_path"
	"github.com/movsb/taoblog/service/modules/renderers/rss"
	"github.com/movsb/taoblog/service/modules/renderers/scoped_css"
	"github.com/movsb/taoblog/service/modules/renderers/stringify"
	"github.com/movsb/taoblog/service/modules/renderers/task_list"
	"github.com/yuin/goldmark/extension"

	wikitable "github.com/movsb/goldmark-wiki-table"
	_ "github.com/movsb/taoblog/theme/share/image_viewer"
	_ "github.com/movsb/taoblog/theme/share/vim"
)

// 发表/更新评论时：普通用户不能发表 HTML 评论，管理员可以。
// 一旦发表/更新成功：始终认为评论是合法的。
//
// 换言之，发表/更新调用此接口，把评论转换成 html 时用 cached 接口。
// 前者用请求身份，后者不限身份。
func (s *Service) renderMarkdown(ctx context.Context, secure bool, postId, _ int64, sourceType, source string, metas models.PostMeta, co *proto.PostContentOptions) (string, error) {
	var tr renderers.Renderer
	switch sourceType {
	case `html`:
		tr = &renderers.HTML{}
		return tr.Render(source)
	case `plain`:
		return html.EscapeString(source), nil
	}

	if sourceType != `markdown` {
		return ``, fmt.Errorf(`unknown source type`)
	}

	options := []renderers.Option2{}
	if postId > 0 {
		if link := s.GetLink(postId); link != s.plainLink(postId) {
			options = append(options, renderers.WithModifiedAnchorReference(link))
		}
		if !co.KeepTitleHeading {
			options = append(options, renderers.WithRemoveTitleHeading())
		}

		options = append(options, media_tags.New(s.OpenAsset(postId)))
		options = append(options, scoped_css.New(fmt.Sprintf(`article.post-%d .entry .content`, postId)))
	}
	if !secure {
		options = append(options,
			renderers.WithDisableHeadings(true),
			renderers.WithDisableHTML(true),
		)
	}
	if co.RenderCodeBlocks {
		options = append(options, highlight.New())
	}
	if co.PrettifyHtml {
		options = append(options, renderers.WithHtmlPrettifier(stringify.New()))
	}
	if co.UseAbsolutePaths {
		options = append(options, rooted_path.New(s.OpenAsset(postId)))
	}
	options = append(options,
		media_size.New(s.OpenAsset(postId),
			media_size.WithLocalOnly(),
			media_size.WithNodeFilter(gold_utils.NegateNodeFilter(withEmojiFilter)),
		),
		image.New(func(path string) (name string, url string, description string, found bool) {
			if src, ok := metas.Sources[path]; ok {
				name = src.Name
				url = src.URL
				description = src.Description
				found = true
			}
			return
		}),
		gallery.New(),
		task_list.New(),
		hashtags.New(s.hashtagResolver, nil),
		custom_break.New(),
		list_markers.New(),
		iframe.New(!co.NoIframePreview),
		math.New(),
		exif.New(s.OpenAsset(postId), s.exifTask, int(postId), exif.WithNodeFilter(gold_utils.NegateNodeFilter(withEmojiFilter))),
		live_photo.New(ctx, s.OpenAsset(postId)),
		emojis.New(emojis.BaseURLForDynamic),
		wikitable.New(),
		extension.GFM,
		footnotes.New(),
		alerts.New(),

		page_link.New(ctx, s.getPostTitle, nil),

		renderers.WithFencedCodeBlockRenderer(`friends`, friends.New(s.friendsTask, int(postId))),
		renderers.WithFencedCodeBlockRenderer(`reminder`, reminders.New()),
		renderers.WithFencedCodeBlockRenderer(`plantuml`, plantuml.NewDefaultSVG(plantuml.WithFileCache(s.fileCache.GetOrLoad))),
		renderers.WithFencedCodeBlockRenderer(`pikchr`, pikchr.New()),
		renderers.WithFencedCodeBlockRenderer(`dot`, graph_viz.New()),
		renderers.WithFencedCodeBlockRenderer(`genealogy`, genealogy.New()),
		renderers.WithFencedCodeBlockRenderer(`rss`, rss.New(s.rssTask, int(postId))),
		renderers.WithFencedCodeBlockRenderer(`echarts`, echarts.New(s.fileCache.GetOrLoad)),

		// 所有人禁止贴无效协议的链接。
		invalid_scheme.New(),

		// 其它选项可能会插入链接，所以放后面。
		// BUG: 放在 html 的最后执行，不然无效，对 hashtags。
		link_target.New(link_target.OpenLinksInNewTabKind(co.OpenLinksInNewTab)),
	)

	tr = renderers.NewMarkdown(options...)
	rendered, err := tr.Render(source)
	if err != nil {
		return "", err
	}

	return rendered, nil
}
