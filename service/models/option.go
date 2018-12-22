package models

import (
	"github.com/movsb/taoblog/protocols"
)

type Option struct {
	ID    string
	Name  string
	Value string
}

func (o *Option) Serialize() *protocols.Option {
	return &protocols.Option{
		Name:  o.Name,
		Value: o.Value,
	}
}

type Options []*Option

func (os Options) Serialize() []*protocols.Option {
	ss := []*protocols.Option{}
	for _, o := range os {
		ss = append(ss, o.Serialize())
	}
	return ss
}
