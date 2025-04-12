package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/movsb/taoblog/protocols/go/proto"
)

const Untitled = `无标题`

type PostStatus string

const (
	PostStatusPublic  = `public`  // 所有人可见。
	PostStatusPrivate = `draft`   // 仅自己可见。
	PostStatusPartial = `partial` // 仅对指定的人可见。
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
	Category        uint
	Status          string
	PageView        uint
	CommentStatus   uint
	Comments        uint
	Metas           PostMeta
	Source          string
	SourceType      string

	DateTimezone     string
	ModifiedTimezone string
}

// NOTE 如果要添加字段，记得同步 isEmpty 方法。
type PostMeta struct {
	Header   string `json:"header,omitempty" yaml:"header,omitempty"`
	Footer   string `json:"footer,omitempty" yaml:"footer,omitempty"`
	Outdated bool   `json:"outdated,omitempty" yaml:"outdated,omitempty"`
	Wide     bool   `json:"wide,omitempty" yaml:"wide,omitempty"`

	Weixin string `json:"weixin,omitempty" yaml:"weixin,omitempty"`

	Sources map[string]*PostMetaSource `json:"sources,omitempty" yaml:"sources,omitempty"`

	Geo *Geo `json:"geo,omitempty" yaml:"geo,omitempty"`

	Origin *proto.Metas_Origin `json:"origin:omitempty" yaml:"origin,omitempty"`

	Toc bool `json:"toc,omitempty" yaml:"toc,omitempty"`
}

type PostMetaSource struct {
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	URL         string `json:"url,omitempty" yaml:"url,omitempty"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Time        int32  `json:"time,omitempty" yaml:"time,omitempty"`
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
		Header:   m.Header,
		Footer:   m.Footer,
		Outdated: m.Outdated,
		Wide:     m.Wide,
		Weixin:   m.Weixin,
		Sources:  make(map[string]*proto.Metas_Source),
		Toc:      m.Toc,
	}
	for name, src := range m.Sources {
		p.Sources[name] = &proto.Metas_Source{
			Name:        src.Name,
			Url:         src.URL,
			Description: src.Description,
			Time:        src.Time,
		}
	}
	if g := m.Geo; g != nil && g.Longitude != 0 && g.Latitude != 0 {
		p.Geo = &proto.Metas_Geo{
			Longitude: g.Longitude,
			Latitude:  g.Latitude,
			Name:      g.Name,
		}
	}
	p.Origin = m.Origin
	return p
}

func (m *PostMeta) IsEmpty() bool {
	return m.Header == "" &&
		m.Footer == "" &&
		!m.Outdated &&
		!m.Wide &&
		m.Weixin == "" &&
		len(m.Sources) == 0 &&
		(m.Geo == nil || (m.Geo.Longitude == 0 && m.Geo.Latitude == 0)) &&
		!m.Toc
}

func PostMetaFrom(p *proto.Metas) *PostMeta {
	if p == nil {
		p = &proto.Metas{}
	}
	m := PostMeta{
		Header:   p.Header,
		Footer:   p.Footer,
		Outdated: p.Outdated,
		Wide:     p.Wide,
		Weixin:   p.Weixin,
		Sources:  make(map[string]*PostMetaSource),
		Toc:      p.Toc,
	}
	for name, src := range p.Sources {
		m.Sources[name] = &PostMetaSource{
			Name:        src.Name,
			URL:         src.Url,
			Description: src.Description,
			Time:        src.Time,
		}
	}
	if g := p.Geo; g != nil {
		m.Geo = &Geo{
			Longitude: g.Longitude,
			Latitude:  g.Latitude,
			Name:      g.Name,
		}
	}
	m.Origin = p.Origin
	return &m
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
		Category:      int64(p.Category),
		Status:        p.Status,
		PageView:      int64(p.PageView),
		CommentStatus: p.CommentStatus > 0,
		Comments:      int64(p.Comments),
		Metas:         p.Metas.ToProto(),
		Source:        p.Source,
		SourceType:    p.SourceType,

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
