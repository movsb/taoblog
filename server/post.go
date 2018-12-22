package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strings"

	"github.com/movsb/taoblog/server/modules/post_translators"
	"github.com/movsb/taoblog/server/modules/utils/datetime"
)

// Post is the model of a post.
type Post struct {
	ID            int64         `json:"id"`
	Date          string        `json:"date"`
	Modified      string        `json:"modified"`
	Title         string        `json:"title"`
	Content       template.HTML `json:"content"`
	Slug          string        `json:"slug"`
	Type          string        `json:"type"`
	Category      uint          `json:"category" taorm:"name:taxonomy"`
	Status        string        `json:"status"`
	PageView      uint          `json:"page_view"`
	CommentStatus uint          `json:"comment_status"`
	Comments      uint          `json:"comments"`
	Metas         string        `json:"metas"`
	Source        string        `json:"source"`
	SourceType    string        `json:"source_type"`
	Tags          []string      `json:"tags"`

	_Metas map[string]interface{}
}

// NewPost news a post object.
func NewPost() *Post {
	p := Post{}
	return &p
}

func (p *Post) PageType() string {
	return "post"
}

func (p *Post) DateString() string {
	d := strings.Split(strings.Split(p.Date, " ")[0], "-")
	return fmt.Sprintf("%v年%v月%v日", d[0], d[1], d[2])
}

func (p *Post) ModifiedString() string {
	d := strings.Split(strings.Split(p.Modified, " ")[0], "-")
	return fmt.Sprintf("%v年%v月%v日", d[0], d[1], d[2])
}

func (p *Post) TagsString() template.HTML {
	var ts []string
	for _, t := range p.Tags {
		et := html.EscapeString(t)
		ts = append(ts, fmt.Sprintf(
			`<a href="/tags/%s">%s</a>`,
			et, et,
		),
		)
	}
	return template.HTML(strings.Join(ts, " · "))
}

func (p *Post) RelatedPosts() []*PostForRelated {
	key := fmt.Sprintf("related:%d", p.ID)
	if r, ok := memcch.Get(key); ok {
		return r.([]*PostForRelated)
	}
	r, err := postmgr.GetRelatedPosts(gdb, p.ID)
	memcch.SetIf(err == nil, key, r)
	return r
}

func (*Post) insertions() string {
	return `date,modified,title,content,slug,type,taxonomy,status,page_view,comment_status,comments,metas,source,source_type`
}
func (*Post) marks() string {
	return `?,?,?,?,?,?,?,?,?,?,?,?,?,?`
}
func (z *Post) updates() string {
	return `date=?,modified=?,title=?,content=?,slug=?,type=?,taxonomy=?,status=?,comment_status=?,metas=?,source=?,source_type=?`
}
func (z *Post) _pointers() []interface{} {
	return []interface{}{
		&z.Date, &z.Modified,
		&z.Title, &z.Content,
		&z.Slug, &z.Type, &z.Category,
		&z.Status, &z.PageView,
		&z.CommentStatus, &z.Comments,
		&z.Metas, &z.Source, &z.SourceType,
	}
}
func (z *Post) values() []interface{} {
	return []interface{}{
		z.Date, z.Modified,
		z.Title, z.Content,
		z.Slug, z.Type, z.Category,
		z.Status, z.PageView,
		z.CommentStatus, z.Comments,
		z.Metas, z.Source, z.SourceType,
	}
}
func (z *Post) update_values() []interface{} {
	return []interface{}{
		z.Date, z.Modified,
		z.Title, z.Content,
		z.Slug, z.Type, z.Category,
		z.Status,
		z.CommentStatus,
		z.Metas, z.Source, z.SourceType,
	}
}

// validate validates fields.
func (z *Post) validate() error {
	if strings.TrimSpace(string(z.Title)) == "" {
		return fmt.Errorf("标题不能为空")
	}
	if z.Content != "" {
		return fmt.Errorf("不能直接指定内容，请指定源")
	}
	if z.Slug != "" {
		if matched, _ := regexp.MatchString(`^[\w-_]+$`, z.Slug); !matched {
			return fmt.Errorf("Slug不规范")
		}
	} else {
		if z.Type == "page" {
			return fmt.Errorf("页面的slug不可以为空")
		}
	}
	// TODO parent page check
	// TODO category existence check
	if z.Category != 0 {
		// return fmt.Errorf("不能指定分类")
	}
	if z.Type == "" {
		z.Type = "post"
	}
	if !strInSlice([]string{"post", "page"}, z.Type) {
		return fmt.Errorf("类型不正确")
	}
	if z.Status == "" {
		z.Status = "public"
	}
	if !strInSlice([]string{"public", "draft"}, z.Status) {
		return fmt.Errorf("发表状态不正确")
	}
	if z.SourceType == "" {
		z.SourceType = "html"
	}
	if !strInSlice([]string{"html", "markdown"}, z.SourceType) {
		return fmt.Errorf("不支持的文章分类：%v", z.SourceType)
	}
	if z.Source == "" {
		return fmt.Errorf("源内容不能为空")
	}
	if z.Date == "" {
		z.Date = datetime.MyLocal()
	}
	if !datetime.IsValidMy(z.Date) {
		return fmt.Errorf("发表时间不正确")
	}
	if z.Modified == "" {
		z.Modified = z.Date
	}
	if !datetime.IsValidMy(z.Modified) {
		return fmt.Errorf("修改时间不正确")
	}
	return nil
}

// translate is called before save into database.
func (z *Post) translate() error {
	var err error

	if z.Category == 0 {
		z.Category = 1
	}
	z.Metas = "{}"
	z.CommentStatus = 1

	z.Date = datetime.Local2My(z.Date)
	z.Modified = datetime.Local2My(z.Modified)

	var tr post_translators.PostTranslator

	switch z.SourceType {
	case "html":
		tr = &post_translators.HTMLTranslator{}
	case "markdown":
		tr = &post_translators.MarkdownTranslator{}
	}

	if tr == nil {
		return errors.New("no translator found for " + z.Type)
	}

	content, err := tr.Translate(z.Source)
	if err != nil {
		return err
	}

	z.Content = template.HTML(content)

	return nil
}

// Create saves post into database.
// It sets ID to lastInsertID.
func (z *Post) Create(tx Querier) error {
	var err error

	if err = z.validate(); err != nil {
		return err
	}
	if err = z.translate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`INSERT INTO posts (%s) VALUES (%s)`, z.insertions(), z.marks())
	ret, err := tx.Exec(query, z.values()...)
	if err != nil {
		return err
	}
	z.ID, err = ret.LastInsertId()
	if err != nil {
		return err
	}

	z.Date = datetime.My2Local(z.Date)
	z.Modified = datetime.My2Local(z.Modified)

	return nil
}

// Update updates a post.
func (z *Post) Update(tx Querier) error {
	var err error

	if err = z.validate(); err != nil {
		return err
	}
	if err = z.translate(); err != nil {
		return err
	}

	query := fmt.Sprintf(`UPDATE posts SET %s WHERE id=%d`, z.updates(), z.ID)
	_, err = tx.Exec(query, z.update_values()...)
	if err != nil {
		return err
	}

	z.Date = datetime.My2Local(z.Date)
	z.Modified = datetime.My2Local(z.Modified)

	return nil
}

func (p *Post) decodeMetas() {
	if p._Metas == nil {
		json.Unmarshal([]byte(p.Metas), &p._Metas)
	}
}

func (p *Post) CustomHeader() (header string) {
	p.decodeMetas()
	if i, ok := p._Metas["header"]; ok {
		if s, ok := i.(string); ok {
			header = s
		}
	}
	return
}

func (p *Post) CustomFooter() (footer string) {
	p.decodeMetas()
	if i, ok := p._Metas["footer"]; ok {
		if s, ok := i.(string); ok {
			footer = s
		}
	}
	return
}
