package protocols

// Tag is a tag.
type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Alias int64  `json:"alias"`
}

// TagWithCount is a tag with associated post count.
type TagWithCount struct {
	Tag
	Count int64 `json:"count"`
}

type ListTagsWithCountRequest struct {
	Limit      int64
	MergeAlias bool
}

type ListTagsWithCountResponse struct {
	Tags []*TagWithCount `json:"tags"`
}
