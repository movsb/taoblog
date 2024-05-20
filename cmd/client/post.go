package client

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	html5 "golang.org/x/net/html"
	field_mask "google.golang.org/protobuf/types/known/fieldmaskpb"
	yaml "gopkg.in/yaml.v2"
)

var (
	errPostInited     = errors.New(`post already initialized, abort`)
	errPostCreated    = errors.New(`post already posted, use update instead`)
	errPostNotCreated = errors.New(`post not created, use create instead`)
)

var sourceNames = []string{
	"README.md",
	"index.md",
	"README.html",
	"index.html",
}

// PostConfig ...
type PostConfig struct {
	ID       int64           `json:"id" yaml:"id"`
	Title    string          `json:"title" yaml:"title"`
	Modified int32           `json:"modified" yaml:"modified"`
	Tags     []string        `json:"tags" yaml:"tags"`
	Metas    models.PostMeta `json:"metas" yaml:"metas"`
	Slug     string          `json:"slug" yaml:"slug,omitempty"`
	Type     string          `json:"type" yaml:"type"`
}

// InitPost ...
func (c *Client) InitPost() error {
	// 禁止意外在项目下创建。
	if _, err := os.Stat(`go.mod`); err == nil {
		log.Fatalln(`不允许在项目根目录下创建文章。`)
	}

	fp, err := os.Open("config.yml")
	if err == nil {
		fp.Close()
		return errPostInited
	}
	fp.Close()
	config := PostConfig{}
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
	p := protocols.Post{}
	cfg := *c.readPostConfig()
	if cfg.ID > 0 {
		return errPostCreated
	}

	p.Title = cfg.Title
	p.Tags = cfg.Tags
	p.Slug = cfg.Slug
	p.Type = cfg.Type
	p.Metas = cfg.Metas.ToProtocols()

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
	post, err := c.Blog.GetPost(c.Context(), &protocols.GetPostRequest{
		Id: int32(cfg.ID),
		ContentOptions: &protocols.PostContentOptions{
			WithContent: false,
		},
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
func (c *Client) SetPostStatus(id int64, public bool, touch bool) {
	if id <= 0 {
		config := c.readPostConfig()
		if config.ID == 0 {
			panic("post not yet been created")
		}
		id = config.ID
	}
	_, err := c.Blog.SetPostStatus(c.Context(), &protocols.SetPostStatusRequest{
		Id:     id,
		Public: public,
		Touch:  touch,
	})
	if err != nil {
		panic(err)
	}
}

// UpdatePost ...
func (c *Client) UpdatePost() error {
	p := protocols.Post{}
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
	p.Metas = cfg.Metas.ToProtocols()
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

	rp, err := c.Blog.UpdatePost(c.Context(), &protocols.UpdatePostRequest{
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
	_, err := c.Blog.DeletePost(c.Context(), &protocols.DeletePostRequest{
		Id: int32(id),
	})
	return err
}

// UploadPostFiles 上传文章附件。
// TODO 应该像 Backup 那样改成带进度的 protocol buffer 方式上传。
// NOTE 路径列表，相对于工作目录，相对路径。
// TODO 由于评论中可能也带有图片引用，但是不会被算计到。所以远端的多余文件总是不会被删除。
// NOTE 会自动去重本地文件。
// NOTE 会自动排除 config.yml 文件。
func (c *Client) UploadPostFiles(id int64, files []string) {
	files = slices.DeleteFunc(files, func(f string) bool { return f == `config.yml` })

	if len(files) <= 0 {
		return
	}

	client, err := c.Management.FileSystem(c.Context())
	if err != nil {
		panic(err)
	}
	defer client.CloseSend()

	fsync := NewFilesSyncer(client)

	localFiles, err := fsync.ListLocalFilesFromPaths(files)
	if err != nil {
		panic(err)
	}

	if err := fsync.SyncPostFiles(id, localFiles); err != nil {
		panic(err)
	}
}

func (c *Client) readPostConfig() *PostConfig {
	config := PostConfig{}
	file, err := os.Open("config.yml")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	dec := yaml.NewDecoder(file)
	if err := dec.Decode(&config); err != nil {
		panic(err)
	}
	return &config
}

func (c *Client) savePostConfig(config *PostConfig) {
	file, err := os.Create("config.yml")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	enc := yaml.NewEncoder(file)
	if err := enc.Encode(config); err != nil {
		panic(err)
	}
}

func readSource(dir string) (string, string, []string) {
	var source string
	var theName string

	for _, name := range sourceNames {
		path := filepath.Join(dir, name)
		bys, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		source = string(bys)
		theName = name
		break
	}
	if source == "" {
		panic("source cannot be found")
	}

	if strings.IndexByte(source, '\x08') != -1 {
		panic("source cannot have '\\x08' characters")
	}
	if strings.Contains(source, "\xe2\x80\x8b") {
		panic("source cannot contain zero width characters")
	}

	typ := ""
	var assets []string
	var err error
	switch filepath.Ext(theName) {
	case ".md":
		typ = "markdown"
		assets, err = parsePostAssets(source)
		if err != nil {
			log.Println(err)
		}
	case ".html":
		typ = "html"
	}

	return typ, source, assets
}

// 从文章的源代码里面提取出附件列表。
// 参考：docs/usage/文章编辑::自动附件上传
// TODO 暂时放在 client 中，其实 server 中也可能用到，到时候再独立成公共模块
// TODO 目前此函数只针对 Markdown 类型的文章，HTML 类型的文章不支持。
func parsePostAssets(source string) ([]string, error) {
	sourceBytes := []byte(source)
	reader := text.NewReader(sourceBytes)
	doc := goldmark.DefaultParser().Parse(reader)

	// 用来保存所有的相对路径列表
	var assets []string

	tryAdd := func(asset string) {
		if strings.Contains(asset, `://`) || !filepath.IsLocal(asset) {
			if asset != "" && (!strings.Contains(asset, `://`) && !filepath.IsAbs(asset)) {
				log.Println(`maybe an invalid asset presents in the post:`, asset)
			}
			return
		}
		// 锚点儿
		if strings.HasPrefix(asset, `#`) {
			return
		}
		// TODO 简单方式去掉 ? 后面的查询参数，有 BUG，但是“够用”。
		asset, _, _ = strings.Cut(asset, "?")
		assets = append(assets, asset)
	}

	fromHTML := func(html string) {
		assets, err := parseHtmlAssets(html)
		if err != nil {
			log.Println(err)
		}
		for _, asset := range assets {
			tryAdd(asset)
		}
	}

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		// 如果修改了这个列表，注意同时更新到文档。
		switch tag := n.(type) {
		case *ast.Link:
			tryAdd(string(tag.Destination))
		case *ast.Image:
			tryAdd(string(tag.Destination))
		case *ast.HTMLBlock, *ast.RawHTML:
			var lines *text.Segments
			switch tag := n.(type) {
			default:
				panic(`unknown tag type`)
			case *ast.HTMLBlock:
				lines = tag.Lines()
			case *ast.RawHTML:
				lines = tag.Segments
			}

			var rawLines []string
			for i := 0; i < lines.Len(); i++ {
				seg := lines.At(i)
				value := seg.Value(sourceBytes)
				rawLines = append(rawLines, string(value))
			}
			fromHTML(strings.Join(rawLines, "\n"))
		}
		return ast.WalkContinue, nil
	})

	return assets, nil
}

func parseHtmlAssets(html string) ([]string, error) {
	node, err := html5.Parse(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var assets []string

	var recurse func(node *html5.Node)

	// 先访问节点自身，再访问各子节点
	recurse = func(node *html5.Node) {
		if !(node.Type == html5.DocumentNode || node.Type == html5.ElementNode) {
			return
		}

		// log.Println("Data:", node.Data)
		var path string
		var wantedAttr string
		switch strings.ToLower(node.Data) {
		case `a`:
			wantedAttr = `href`
		case `img`, `source`, `iframe`:
			wantedAttr = `src`
		case `object`:
			wantedAttr = `data`
		}
		if wantedAttr != `` {
			for _, attr := range node.Attr {
				if strings.EqualFold(attr.Key, wantedAttr) {
					path = attr.Val
				}
			}
		}
		if path != `` {
			assets = append(assets, path)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			recurse(child)
		}
	}

	recurse(node)

	return assets, nil
}
