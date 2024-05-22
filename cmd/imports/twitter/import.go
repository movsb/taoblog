package twitter

import (
	"bytes"
	"html/template"
	"io/fs"
	"log"
	"strings"

	"github.com/movsb/taoblog/cmd/client"
	proto "github.com/movsb/taoblog/protocols"
)

type Importer struct {
	root   fs.FS
	client *proto.ProtoClient
}

func New(root fs.FS, client *proto.ProtoClient) *Importer {
	return &Importer{
		root:   root,
		client: client,
	}
}

// TODO 允许增量导入（对存在相同ID的推文只创建一条对应的碎碎念）。
func (i *Importer) Execute() error {
	tweets, err := ParseTweets(i.root)
	if err != nil {
		return err
	}

	// 为了上传附件，切到附件目录。
	tweetsMedia, err := fs.Sub(i.root, `data/tweets_media`)
	if err != nil {
		return err
	}

	for _, tweet := range tweets {
		// 先只考虑自己发的（不包含回复、转推）。
		if !tweet.IsSelfTweet() {
			continue
		}

		post, err := convertToPost(tweet)
		if err != nil {
			log.Fatalln(err)
		}
		post, err = i.client.Blog.CreatePost(i.client.Context(), post)
		if err != nil {
			log.Fatalln(err)
		}

		images, videos := tweet.Assets(true)
		client.UploadPostFiles(i.client, post.Id, tweetsMedia, images)
		videoNames := make([]string, 0, len(videos))
		for _, v := range videos {
			videoNames = append(videoNames, v.FileName)
		}
		client.UploadPostFiles(i.client, post.Id, tweetsMedia, videoNames)
	}

	return nil
}

func convertToPost(t *Tweet) (*proto.Post, error) {
	p := proto.Post{
		Date: int32(t.CreatedAt),

		// 普通用户没有编辑权限，直接用创建时间。
		Modified: int32(t.CreatedAt),

		// 后台自动生成
		Title: "",
		// 原文是普通纯文本。可能带 < >
		SourceType: `markdown`,
		Source:     string(t.Markdown()),

		Type:   `tweet`,
		Status: `public`,

		Metas: &proto.Metas{
			Origin: &proto.Metas_Origin{
				Platform: proto.Metas_Origin_Twitter,
				// 应该不会导入别人的吧？
				// 所以只保留了推文本身的编号，而不包含用户名。
				// 而且，推特的唯一性是 ID，根用户名无关。
				// https://stackoverflow.com/a/27843083/3628322
				// https://developer.x.com/en/blog/community/2020/getting-to-the-canonical-url-for-a-post
				Slugs: []string{t.ID},
			},
		},

		Tags: t.TagNames(),
	}

	useMedia(&p.Source, t)

	// 为简单起见，评论直接嵌入原推。
	var recurse func(t *Tweet)
	recurse = func(t *Tweet) {
		p.Source += "\n\n<hr>\n\n"
		p.Source += t.Markdown()
		useMedia(&p.Source, t)

		for _, c := range t.children {
			recurse(c)
		}
	}

	for _, c := range t.children {
		recurse(c)
	}

	return &p, nil
}

// func convertToComment(t *Tweet) (*proto.Comment, error) {
// 	c := proto.Comment{

// 	}
// 	useMedia(&c.Source, t)
// 	return &c, nil
// }

func useMedia(source *string, t *Tweet) {
	// 如果有附件，把它们拼接到内容后面。
	// 需要前面的内容是 markdown 或者 html。
	images, videos := t.Assets(false)
	if len(images) > 0 || len(videos) > 0 {
		var classNames []string
		if len(images) > 0 {
			classNames = append(classNames, `has-images`)
		}
		if len(videos) > 0 {
			classNames = append(classNames, `has-videos`)
		}
		if len(images)+len(videos) > 1 {
			classNames = append(classNames, `has-multiple-media`)
		}

		data := struct {
			Images     []string
			Videos     []VideoAsset
			ClassNames string
		}{
			Images:     images,
			Videos:     videos,
			ClassNames: strings.Join(classNames, ` `),
		}
		buf := bytes.NewBuffer(nil)
		if err := mediaTmpl.Execute(buf, data); err != nil {
			log.Fatalln(err)
		}

		*source += "\n\n"
		*source += buf.String()
	}
}

// 保证在即便不加载脚本的情况下也能正确渲染，这样对搜索引擎友好
// 注意别被缩进渲染成代码。
// NOTE：对 video 的 media_url 不包含在数据目录中，无法用作 video 的 poster。
// NOTE：图片写作 markdown 形式是为了渲染的时候自动计算大小和 lazying。
// TODO：把上述计算转移到 html filter 里面做。
// TODO 处理 poster & preload。
// NOTE: 注意把图片渲染在一个 <p> 内，方便九宫格排版。
var mediaTmpl = template.Must(template.New(`media`).Parse(`
<div class="tweet-media {{.ClassNames}}">

{{range .Images }}
![{{.}}]({{.}})
{{- end }}

{{- range .Videos }}
<video controls {{ if .Width}}width="{{.Width}}" height="{{.Height}}"{{end}} data-poster="{{.PosterURL}}" />
<source src="{{.FileName}}" type="{{.ContentType}}"/>
</video>
{{- end }}
</div>
`))
