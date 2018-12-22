package protocols

type Rss struct {
	Content  string `json:"content"`
	Modified string `json:"modified"`
}

type GetRssRequest struct {
}
