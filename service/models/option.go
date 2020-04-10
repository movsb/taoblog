package models

type Option struct {
	ID    int64
	Name  string
	Value string
}

// TableName ...
func (Option) TableName() string {
	return `options`
}
