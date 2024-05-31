package client

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
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

const ConfigFileName = `config.yml`

// InitPost ...
func (c *Client) InitPost() error {
	// 禁止意外在项目下创建。
	if _, err := os.Stat(`go.mod`); err == nil {
		log.Fatalln(`不允许在项目根目录下创建文章。`)
	}

	fp, err := os.Open(ConfigFileName)
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
		Id: int32(cfg.ID),
		ContentOptions: &proto.PostContentOptions{
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
	_, err := c.Blog.SetPostStatus(c.Context(), &proto.SetPostStatusRequest{
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
	files = slices.DeleteFunc(files, func(f string) bool { return f == ConfigFileName })

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
		panic(err)
	}
}

func (c *Client) readPostConfig() *PostConfig {
	p, err := ReadPostConfig(ConfigFileName)
	if err != nil {
		panic(err)
	}
	return p
}

func ReadPostConfig(path string) (*PostConfig, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return ReadPostConfigReader(file)
}
func ReadPostConfigReader(r io.Reader) (*PostConfig, error) {
	config := PostConfig{}
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Client) savePostConfig(config *PostConfig) {
	if err := SavePostConfig(ConfigFileName, config); err != nil {
		panic(err)
	}
}

func SavePostConfig(path string, config *PostConfig) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := yaml.NewEncoder(file)
	if err := enc.Encode(config); err != nil {
		return err
	}
	return nil
}

const IndexFileName = `README.md`

func readSource(dir string) (string, string, []string) {
	var source string

	path := filepath.Join(dir, IndexFileName)
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
// TODO 目前此函数只针对 Markdown 类型的文章，HTML 类型的文章不支持。
func parsePostAssets(source string) ([]string, error) {
	sourceBytes := []byte(source)
	reader := text.NewReader(sourceBytes)
	doc := goldmark.DefaultParser().Parse(reader)

	// 用来保存所有的相对路径列表
	var assets []string

	tryAdd := func(theURL string) {
		u, err := url.Parse(theURL)
		if err != nil {
			log.Println(err)
			return
		}
		if u.Scheme != "" || u.Host != "" || strings.HasPrefix(u.Path, `/`) {
			log.Println(`maybe an invalid asset presents in the post:`, theURL)
			return
		}
		relative := u.Path
		// 锚点儿
		if strings.HasPrefix(relative, `#`) {
			return
		}
		// TODO 简单方式去掉 ? 后面的查询参数，有 BUG，但是“够用”。
		relative, _, _ = strings.Cut(relative, "?")
		assets = append(assets, relative)
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

		var theURL string
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
					theURL = attr.Val
				}
			}
		}
		if theURL != `` {
			assets = append(assets, theURL)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			recurse(child)
		}
	}

	recurse(node)

	return assets, err
}
