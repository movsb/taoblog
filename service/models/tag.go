package models

import "github.com/movsb/taoblog/protocols/go/proto"

type Tag struct {
	ID    int64
	Name  string
	Alias int64
}

func (Tag) TableName() string {
	return `tags`
}

type ObjectTag struct {
	ID     int64
	PostID int64
	TagID  int64
}

func (ObjectTag) TableName() string {
	return `post_tags`
}

type Category struct {
	ID       int32
	UserID   int32
	ParentID int32
	Name     string
}

func (Category) TableName() string {
	return `categories`
}

func (c *Category) ToProto() (*proto.Category, error) {
	return &proto.Category{
		Id:       c.ID,
		UserId:   c.UserID,
		ParentId: c.ParentID,
		Name:     c.Name,
	}, nil
}

type Categories []*Category

func (cats Categories) ToProto() ([]*proto.Category, error) {
	out := make([]*proto.Category, 0, len(cats))
	for _, c := range cats {
		c1, err := c.ToProto()
		if err != nil {
			return nil, err
		}
		out = append(out, c1)
	}
	return out, nil
}
