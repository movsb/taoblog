package client

import (
	"bytes"
	"errors"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/PuerkitoBio/goquery"
	client_common "github.com/movsb/taoblog/cmd/client/common"
	"github.com/movsb/taoblog/protocols/clients"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	field_mask "google.golang.org/protobuf/types/known/fieldmaskpb"
)

var (
	errPostInited     = errors.New(`post already initialized, abort`)
	errPostCreated    = errors.New(`post already posted, use update instead`)
	errPostNotCreated = errors.New(`post not created, use create instead`)
)

// InitPost ...
func (c *Client) InitPost() error {
	// 禁止意外在项目下创建。
	if _, err := os.Stat(`go.mod`); err == nil {
		log.Fatalln(`不允许在项目根目录下创建文章。`)
	}

	fp, err := os.Open(client_common.ConfigFileName)
	if err == nil {
		fp.Close()
		return errPostInited
	}
	fp.Close()
	config := client_common.PostConfig{}
	c.savePostConfig(&config)
	// try to touch README.md
	fp, _ = os.OpenFile("README.md", os.O_RDONLY|os.O_CREATE, 0644)
	if fp != nil {
		fp.Close()
	}
	return nil
}

// CreatePost ...
func (c *Client) CreatePost() error {
	p := proto.Post{}
	cfg := *c.readPostConfig()
	if cfg.ID > 0 {
		return errPostCreated
	}

	p.Title = cfg.Title
	p.Tags = cfg.Tags
	p.Slug = cfg.Slug
	p.Type = cfg.Type
	p.Metas = cfg.Metas.ToProto()

	if p.Type == "" {
		p.Type = `post`
	}

	var assets []string

	p.SourceType, p.Source, assets = readSource(".")

	rp, err := c.Blog.CreatePost(c.Context(), &p)
	if err != nil {
		return err
	}

	cfg.ID = rp.Id
	cfg.Modified = rp.Modified
	c.savePostConfig(&cfg)

	// TODO 应该先上传文件，但是会拿不到编号
	c.UploadPostFiles(cfg.ID, assets)

	return nil
}

// GetPost ...
func (c *Client) GetPost() {
	cfg := *c.readPostConfig()
	if cfg.ID == 0 {
		panic("ID cannot be zero")
	}
	if cfg.Title != "" {
		panic("must not be created")
	}
	post, err := c.Blog.GetPost(c.Context(), &proto.GetPostRequest{
		Id:             int32(cfg.ID),
		ContentOptions: co.For(co.ClientGetPost),
	})
	if err != nil {
		panic(err)
	}

	cfg.Slug = post.Slug
	cfg.Tags = post.Tags
	cfg.Title = post.Title
	cfg.Modified = post.Modified
	cfg.Type = post.Type
	cfg.Metas = *models.PostMetaFrom(post.Metas)
	c.savePostConfig(&cfg)

	filename := "README.md"
	switch post.SourceType {
	case "html":
		filename = "README.html"
		if post.Source == "" {
			post.Source = post.Content
		}
	case "markdown":
		filename = "README.md"
	}
	fp, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	if _, err := fp.WriteString(post.Source); err != nil {
		panic(err)
	}
}

// SetPostStatus ...
func (c *Client) SetPostStatus(id int64, status models.PostStatus, touch bool) {
	if id <= 0 {
		config := c.readPostConfig()
		if config.ID == 0 {
			panic("post not yet been created")
		}
		id = config.ID
	}
	_, err := c.Blog.SetPostStatus(c.Context(), &proto.SetPostStatusRequest{
		Id:     id,
		Status: string(status),
		Touch:  touch,
	})
	if err != nil {
		panic(err)
	}
}

// UpdatePost ...
func (c *Client) UpdatePost() error {
	p := proto.Post{}
	cfg := *c.readPostConfig()
	if cfg.ID == 0 {
		return errPostNotCreated
	}

	p.Id = cfg.ID
	p.Title = cfg.Title
	p.Tags = cfg.Tags
	p.Slug = cfg.Slug
	p.Modified = cfg.Modified
	p.Type = cfg.Type
	p.Metas = cfg.Metas.ToProto()
	if p.Type == "" {
		p.Type = `post`
	}

	var assets []string

	p.SourceType, p.Source, assets = readSource(".")

	updateMasks := []string{
		`title`,
		`source_type`,
		`source`,
		`slug`,
		`tags`,
		`type`,
	}

	// 有些文章在远程有 meta 但是本地没有。
	// 如果更新，则不应该带 meta，而是等待同步回来。
	if !cfg.Metas.IsEmpty() {
		updateMasks = append(updateMasks, `metas`)
	}

	rp, err := c.Blog.UpdatePost(c.Context(), &proto.UpdatePostRequest{
		Post: &p,
		UpdateMask: &field_mask.FieldMask{
			Paths: updateMasks,
		},
	})
	if err != nil {
		return err
	}
	cfg.Title = rp.Title
	cfg.Tags = rp.Tags
	cfg.Slug = rp.Slug
	cfg.Modified = rp.Modified
	cfg.Type = rp.Type
	cfg.Metas = *models.PostMetaFrom(rp.Metas)
	c.savePostConfig(&cfg)

	// TODO 应该先上传文件，但是会拿不到编号
	c.UploadPostFiles(cfg.ID, assets)

	return nil
}

// DeletePost ...
func (c *Client) DeletePost(id int64) error {
	_, err := c.Blog.DeletePost(c.Context(), &proto.DeletePostRequest{
		Id: int32(id),
	})
	return err
}

func (c *Client) UploadPostFiles(id int64, files []string) {
	UploadPostFiles(c.ProtoClient, id, os.DirFS("."), files)
}

// UploadPostFiles 上传文章附件。
// TODO 应该像 Backup 那样改成带进度的 protocol buffer 方式上传。
// NOTE 路径列表，相对于工作目录，相对路径。
// TODO 由于评论中可能也带有图片引用，但是不会被算计到。所以远端的多余文件总是不会被删除。
// NOTE 会自动去重本地文件。
// NOTE 会自动排除 config.yml 文件。
func UploadPostFiles(client *clients.ProtoClient, id int64, root fs.FS, files []string) {
	files = slices.DeleteFunc(files, func(f string) bool { return f == client_common.ConfigFileName })

	if len(files) <= 0 {
		return
	}

	manage, err := client.Management.FileSystem(client.Context())
	if err != nil {
		panic(err)
	}
	defer manage.CloseSend()

	fsync := NewFilesSyncer(manage)

	if err := fsync.SyncPostFiles(id, root, files); err != nil {
		log.Fatalln(err)
	}
}

func (c *Client) readPostConfig() *client_common.PostConfig {
	p, err := client_common.ReadPostConfig(client_common.ConfigFileName)
	if err != nil {
		panic(err)
	}
	return p
}

func (c *Client) savePostConfig(config *client_common.PostConfig) {
	if err := client_common.SavePostConfig(client_common.ConfigFileName, config); err != nil {
		panic(err)
	}
}

func readSource(dir string) (string, string, []string) {
	var source string

	path := filepath.Join(dir, client_common.IndexFileName)
	bys, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	source = string(bys)

	if source == "" {
		panic("source cannot be found")
	}

	if strings.IndexByte(source, '\x08') != -1 {
		panic("source cannot have '\\x08' characters")
	}
	if strings.Contains(source, "\xe2\x80\x8b") {
		panic("source cannot contain zero width characters")
	}

	var assets []string
	typ := "markdown"
	assets, err = parsePostAssets(source)
	if err != nil {
		log.Println(err)
	}

	return typ, source, assets
}

// 从文章的源代码里面提取出附件列表。
// 参考：docs/usage/文章编辑::自动附件上传
// TODO 暂时放在 client 中，其实 server 中也可能用到，到时候再独立成公共模块
func parsePostAssets(source string) ([]string, error) {
	buf := bytes.NewBuffer(nil)
	if err := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	).Convert([]byte(source), buf); err != nil {
		return nil, err
	}
	markup := buf.Bytes()
	_ = markup
	doc, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return nil, err
	}

	// 用来保存所有的相对路径列表
	var assets []string

	tryAdd := func(theURL string) {
		u, err := url.Parse(theURL)
		if err != nil {
			log.Println(err)
			return
		}
		if u.Scheme != "" || u.Host != "" || strings.HasPrefix(u.Path, `/`) {
			// log.Println(`maybe an invalid asset presents in the post:`, theURL)
			return
		}
		if u.Path == "" {
			return
		}
		assets = append(assets, u.Path)
	}

	doc.Find(`a,img,iframe,source,audio,video,object`).Each(func(_ int, s *goquery.Selection) {
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
		tryAdd(attrValue)
	})

	return assets, nil
}
