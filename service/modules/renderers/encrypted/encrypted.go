package encrypted

import (
	"embed"

	"github.com/PuerkitoBio/goquery"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/dir"
	"github.com/movsb/taoblog/service/modules/dynamic"
	"golang.org/x/net/html/atom"
)

func init() {
	dynamic.RegisterInit(func() {
		const module = `encrypted`
		dynamic.WithRoots(module, nil, nil, _embed, _local)
		dynamic.WithScripts(module, `script.js`)
	})
}

//go:embed script.js
var _embed embed.FS
var _local = utils.NewOSDirFS(string(dir.SourceAbsoluteDir()))

type _Encrypted struct{}

func New() *_Encrypted {
	return &_Encrypted{}
}

func (m *_Encrypted) TransformHtml(doc *goquery.Document) error {
	doc.Find(`img,video`).Each(func(i int, s *goquery.Selection) {
		if s.HasClass(`static`) {
			return
		}
		s.SetAttr(`onerror`, `decryptFile(this)`)

		// 场景：文章视频刚上传时，视频在服务器本身，能直接加载成功。当切换到加密源后，因为 fucking 缓存的原因，
		// 无论怎么 video.load()，火狐都不会重新加载视频；但是 fetch 到的确实是 json，Firefox sucks. 表现为，实况照片
		// 再也播放不了了，无论怎么禁用缓存、无论怎么强制刷新。“缓存”真 jb 烦。
		// https://www.reddit.com/r/firefox/comments/13keipv/why_is_a_network_request_showing_up_as_blocked_in/
		//
		// 临时解决办法：监听视频的 `ended` 事件，如果 ended 后 onerror 还存在，那一定是出问题了。
		// 并尝试修复，修复方法参考：https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/Cache-Control#caching_static_assets_with_cache_busting
		if s.Nodes[0].DataAtom == atom.Video {
			s.SetAttr(`onended`, `fixVideoCache(this)`)
		}
	})
	return nil
}
