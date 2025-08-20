package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/movsb/taoblog/protocols/go/proto"
)

const (
	Untitled               = `无标题`
	UntitledSourceMarkdown = `# ` + Untitled + "\n\n"
)

type PostStatus string

const (
	PostStatusPublic  = `public`  // 所有人可见。
	PostStatusPartial = `partial` // 仅对指定的人可见。
	PostStatusPrivate = `private` // 仅自己可见。
	PostStatusDraft   = `draft`   // 仅自己可见，草稿。仅显示在草稿箱，不在文章列表中显示。
)

type Post struct {
	ID              int64
	UserID          int32
	Date            int32
	Modified        int32
	LastCommentedAt int32
	Title           string
	Slug            string
	Type            string
	Category        int32
	Status          string
	PageView        uint
	CommentStatus   uint
	Comments        uint
	Metas           PostMeta
	Source          string
	SourceType      string

	Citations        References
	DateTimezone     string
	ModifiedTimezone string
}

// NOTE 如果要添加字段，记得同步 isEmpty 方法。
type PostMeta struct {
	Header   string `json:"header,omitempty" yaml:"header,omitempty"`
	Footer   string `json:"footer,omitempty" yaml:"footer,omitempty"`
	Outdated bool   `json:"outdated,omitempty" yaml:"outdated,omitempty"`
	Wide     bool   `json:"wide,omitempty" yaml:"wide,omitempty"`
	Toc      bool   `json:"toc,omitempty" yaml:"toc,omitempty"`

	Geo    *Geo                `json:"geo,omitempty" yaml:"geo,omitempty"`
	Origin *proto.Metas_Origin `json:"origin:omitempty" yaml:"origin,omitempty"`

	Weixin     string `json:"weixin,omitempty" yaml:"weixin,omitempty"`
	TextIndent bool   `json:"text_indent,omitempty" yaml:"text_indent,omitempty"`
}

// 本来想用 GeoJSON 的，但是感觉标准化程度还不高。
// 想想还是算了，我只是想不丢失早期说说的地理位置信息。
// 那就用最简单的方式：经、纬度、名字，后期再升级吧。
// https://geojson.org/
// https://en.wikipedia.org/wiki/GeoJSON
type Geo struct {
	Name      string  `json:"name,omitempty" yaml:"name,omitempty"`           // 地理位置的名字
	Longitude float32 `json:"longitude,omitempty" yaml:"longitude,omitempty"` // 经度
	Latitude  float32 `json:"latitude,omitempty" yaml:"latitude,omitempty"`   // 纬度
	Private   bool    `json:"private,omitempty" yaml:"private,omitempty"`     // 是否私有地址
}

var (
	_ sql.Scanner   = (*PostMeta)(nil)
	_ driver.Valuer = (*PostMeta)(nil)
)

func (m PostMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *PostMeta) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	}
	return errors.New(`unsupported type`)
}

func (m *PostMeta) ToProto() *proto.Metas {
	p := &proto.Metas{
		Header:     m.Header,
		Footer:     m.Footer,
		Outdated:   m.Outdated,
		Wide:       m.Wide,
		Weixin:     m.Weixin,
		Toc:        m.Toc,
		TextIndent: m.TextIndent,
	}
	if g := m.Geo; g != nil {
		p.Geo = &proto.Metas_Geo{
			Longitude: g.Longitude,
			Latitude:  g.Latitude,
			Name:      g.Name,
			Private:   g.Private,
		}
	}
	p.Origin = m.Origin
	return p
}

// 不重要的函数。可以删除。加字段可以不更新。
func (m *PostMeta) IsEmpty() bool {
	return m.Header == "" &&
		m.Footer == "" &&
		!m.Outdated &&
		!m.Wide &&
		m.Weixin == "" &&
		(m.Geo == nil || (m.Geo.Longitude == 0 && m.Geo.Latitude == 0)) &&
		!m.Toc
}

func PostMetaFrom(p *proto.Metas) *PostMeta {
	if p == nil {
		p = &proto.Metas{}
	}
	m := PostMeta{
		Header:     p.Header,
		Footer:     p.Footer,
		Outdated:   p.Outdated,
		Wide:       p.Wide,
		Weixin:     p.Weixin,
		Toc:        p.Toc,
		TextIndent: p.TextIndent,
	}
	if g := p.Geo; g != nil {
		m.Geo = &Geo{
			Longitude: g.Longitude,
			Latitude:  g.Latitude,
			Name:      g.Name,
			Private:   g.Private,
		}
	}
	m.Origin = p.Origin
	return &m
}

type References proto.Post_References

var (
	_ sql.Scanner   = (*References)(nil)
	_ driver.Valuer = (*References)(nil)
)

// TODO taorm 有个bug，这里必须为值，不能是指针。
func (m References) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *References) Scan(value any) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	}
	return errors.New(`unsupported type`)
}

func (Post) TableName() string {
	return `posts`
}

// 以下字段由 setPostExtraFields 提供/清除。
// - Metas.Geo
// - Content
// - Tags
func (p *Post) ToProto(redact func(p *proto.Post) error) (*proto.Post, error) {
	out := proto.Post{
		Id:            p.ID,
		UserId:        p.UserID,
		Date:          p.Date,
		Modified:      p.Modified,
		Title:         p.Title,
		Slug:          p.Slug,
		Type:          p.Type,
		Category:      p.Category,
		Status:        p.Status,
		PageView:      int64(p.PageView),
		CommentStatus: p.CommentStatus > 0,
		Comments:      int64(p.Comments),
		Metas:         p.Metas.ToProto(),
		Source:        p.Source,
		SourceType:    p.SourceType,
		References:    (*proto.Post_References)(&p.Citations),

		LastCommentedAt: p.LastCommentedAt,

		DateTimezone:     p.DateTimezone,
		ModifiedTimezone: p.ModifiedTimezone,
	}
	err := redact(&out)
	return &out, err
}

type Posts []*Post

func (ps Posts) ToProto(redact func(p *proto.Post) error) (posts []*proto.Post, err error) {
	for _, post := range ps {
		if p, err := post.ToProto(redact); err != nil {
			return nil, err
		} else {
			posts = append(posts, p)
		}
	}
	return
}
