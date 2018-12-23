package models

import (
	"github.com/movsb/taoblog/protocols"
)

// Tag is a tag.
type Tag struct {
	ID    int64
	Name  string
	Alias int64
}

func (t *Tag) Serialize() *protocols.Tag {
	return &protocols.Tag{
		ID:    t.ID,
		Name:  t.Name,
		Alias: t.Alias,
	}
}

// TagWithCount is a tag with associated post count.
type TagWithCount struct {
	Tag
	Count int64
}

func (t *TagWithCount) Serialize() *protocols.TagWithCount {
	return &protocols.TagWithCount{
		Tag:   *t.Tag.Serialize(),
		Count: t.Count,
	}
}

type TagWithCounts []*TagWithCount

func (ts TagWithCounts) Serialize() []*protocols.TagWithCount {
	st := []*protocols.TagWithCount{}
	for _, t := range ts {
		st = append(st, t.Serialize())
	}
	return st
}
