package client

import (
	"github.com/movsb/taoblog/protocols/go/proto"
)

// SetCommentPostID ...
func (c *Client) SetCommentPostID(commentID int64, postID int64) {
	req := proto.SetCommentPostIDRequest{
		Id:     commentID,
		PostId: postID,
	}
	resp, err := c.Blog.SetCommentPostID(c.Context(), &req)
	if err != nil {
		panic(err)
	}
	_ = resp
}
