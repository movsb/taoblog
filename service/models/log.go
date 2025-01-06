package models

// 代表一条日志或需要处理的事件。
type Log struct {
	ID      int64
	Time    int64
	Type    string
	SubType string
	Version int
	Data    string
}

func (Log) TableName() string {
	return `logs`
}
