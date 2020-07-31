package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/movsb/taoblog/protocols"
	"google.golang.org/genproto/protobuf/field_mask"
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

type PostConfig struct {
	ID    int64    `json:"id" yaml:"id"`
	Title string   `json:"title" yaml:"title"`
	Tags  []string `json:"tags" yaml:"tags"`
	Slug  string   `json:"slug" yaml:"slug,omitempty"`
}

// InitPost ...
func (c *Client) InitPost() error {
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

	p.SourceType, p.Source = readSource(".")

	rp, err := c.grpcClient.CreatePost(c.token(), &p)
	if err != nil {
		return err
	}

	cfg.ID = rp.Id
	c.savePostConfig(&cfg)

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
	post, err := c.grpcClient.GetPost(c.token(), &protocols.GetPostRequest{
		Id:          int32(cfg.ID),
		WithSource:  true,
		WithContent: true,
		WithTags:    true,
	})
	if err != nil {
		panic(err)
	}

	cfg.Slug = post.Slug
	cfg.Tags = post.Tags
	cfg.Title = post.Title
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

func (c *Client) SetPostStatus(id int64, public bool, touch bool) {
	if id <= 0 {
		config := c.readPostConfig()
		if config.ID == 0 {
			panic("post not yet been created")
		}
		id = config.ID
	}
	_, err := c.grpcClient.SetPostStatus(c.token(), &protocols.SetPostStatusRequest{
		Id:     id,
		Public: public,
		Touch:  touch,
	})
	if err != nil {
		panic(err)
	}
}

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

	p.SourceType, p.Source = readSource(".")

	rp, err := c.grpcClient.UpdatePost(c.token(), &protocols.UpdatePostRequest{
		Post: &p,
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{
				`title`,
				`source_type`,
				`source`,
				`slug`,
				`tags`,
			},
		},
	})
	if err != nil {
		panic(err)
	}
	cfg.Title = rp.Title
	cfg.Tags = rp.Tags
	cfg.Slug = rp.Slug
	c.savePostConfig(&cfg)

	return nil
}

// UploadPostFiles ...
func (c *Client) UploadPostFiles(files []string) {
	config := c.readPostConfig()
	if config.ID <= 0 {
		panic("post not posted, post it first.")
	}
	if len(files) <= 0 {
		panic("Specify files.")
	}
	for _, file := range files {
		fmt.Println("  +", file)
		var err error
		fp, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		path := fmt.Sprintf("/posts/%d/files/%s", config.ID, file)
		resp := c.mustPost(path, fp, contentTypeBinary)
		_ = resp
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

func readSource(dir string) (string, string) {
	var source string
	var theName string

	for _, name := range sourceNames {
		path := filepath.Join(dir, name)
		bys, err := ioutil.ReadFile(path)
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

	typ := ""
	switch filepath.Ext(theName) {
	case ".md":
		typ = "markdown"
	case ".html":
		typ = "html"
	}

	return typ, source
}
