package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/movsb/taoblog/protocols"
)

// SetCommentPostID ...
func (c *Client) SetCommentPostID(commentID int64, postID int64) {
	cmt := protocols.Comment{
		ID:     commentID,
		PostID: postID,
	}
	bys, err := json.Marshal(cmt)
	if err != nil {
		panic(err)
	}
	resp := c.mustPost(fmt.Sprintf(`/comments!setPostID`), bytes.NewReader(bys), contentTypeJSON)
	defer resp.Body.Close()
}
