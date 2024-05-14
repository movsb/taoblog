package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/movsb/taoblog/protocols"
)

const Untitled = `无标题`

// 文章的种类。
// 其实就是 type，准备换个名字。
type Kind string

// Post ...
type Post struct {
	ID              int64
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
}

type PostMeta struct {
	Header   string `json:"header,omitempty" yaml:"header,omitempty"`
	Footer   string `json:"footer,omitempty" yaml:"footer,omitempty"`
	Outdated bool   `json:"outdated,omitempty" yaml:"outdated,omitempty"`
	Wide     bool   `json:"wide,omitempty" yaml:"wide,omitempty"`

	Weixin string `json:"weixin,omitempty" yaml:"weixin,omitempty"`

	Sources map[string]*PostMetaSource `json:"sources,omitempty" yaml:"sources,omitempty"`

	Geo Geo `json:"geo,omitempty" yaml:"geo,omitempty"`
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

func (m *PostMeta) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), m)
	case []byte:
		return json.Unmarshal(v, m)
	}
	return errors.New(`unsupported type`)
}

func (m *PostMeta) ToProtocols() *protocols.Metas {
	p := &protocols.Metas{
		Header:   m.Header,
		Footer:   m.Footer,
		Outdated: m.Outdated,
		Wide:     m.Wide,
		Weixin:   m.Weixin,
		Sources:  make(map[string]*protocols.Metas_Source),
	}
	for name, src := range m.Sources {
		p.Sources[name] = &protocols.Metas_Source{
			Name:        src.Name,
			Url:         src.URL,
			Description: src.Description,
			Time:        src.Time,
		}
	}
	if g := m.Geo; g.Longitude != 0 && g.Latitude != 0 {
		p.Geo = &protocols.Metas_Geo{
			Longitude: g.Longitude,
			Latitude:  g.Latitude,
			Name:      g.Name,
		}
	}
	return p
}

func (m *PostMeta) IsEmpty() bool {
	return m.Header == "" &&
		m.Footer == "" &&
		!m.Outdated &&
		!m.Wide &&
		m.Weixin == "" &&
		len(m.Sources) == 0 &&
		m.Geo.Longitude == 0 && m.Geo.Latitude == 0
}

func PostMetaFrom(p *protocols.Metas) *PostMeta {
	if p == nil {
		p = &protocols.Metas{}
	}
	m := PostMeta{
		Header:   p.Header,
		Footer:   p.Footer,
		Outdated: p.Outdated,
		Wide:     p.Wide,
		Weixin:   p.Weixin,
		Sources:  make(map[string]*PostMetaSource),
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
		m.Geo.Longitude = g.Longitude
		m.Geo.Latitude = g.Latitude
		m.Geo.Name = g.Name
	}
	return &m
}

// TableName ...
func (Post) TableName() string {
	return `posts`
}

// 以下字段由 setPostExtraFields 提供/清除。
// - Metas.Geo
// - Content
func (p *Post) ToProtocols(redact func(p *protocols.Post) error) (*protocols.Post, error) {
	out := protocols.Post{
		Id:            p.ID,
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
		Metas:         p.Metas.ToProtocols(),
		Source:        p.Source,
		SourceType:    p.SourceType,

		LastCommentedAt: p.LastCommentedAt,
	}
	err := redact(&out)
	return &out, err
}

// Posts ...
type Posts []*Post

// ToProtocols ...
func (ps Posts) ToProtocols(redact func(p *protocols.Post) error) (posts []*protocols.Post, err error) {
	for _, post := range ps {
		if p, err := post.ToProtocols(redact); err != nil {
			return nil, err
		} else {
			posts = append(posts, p)
		}
	}
	return
}

type PostForRelated struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Relevance uint   `json:"relevance"`
}

type Redirect struct {
	ID         int64
	CreatedAt  int32
	SourcePath string
	TargetPath string
	StatusCode int
}

func (Redirect) TableName() string {
	return `redirects`
}
