package client

import (
	"github.com/movsb/taoblog/protocols"
	field_mask "google.golang.org/protobuf/types/known/fieldmaskpb"
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

func (c *Client) GetComment(cmdID int64) *proto.Comment {
	cmt, err := c.Blog.GetComment(c.Context(), &proto.GetCommentRequest{
		Id: cmdID,
	})
	if err != nil {
		panic(err)
	}
	return cmt
}

// 更新一条评论。
// 非 Markdown 评论会被转换为 Markdown。
func (c *Client) UpdateComment(cmtID int64) {
	cmt := c.GetComment(cmtID)

	value, ok := edit(cmt.Source, `.md`)
	if !ok {
		return
	}

	cmt.SourceType = `markdown`
	cmt.Source = string(value)

	_, err := c.Blog.UpdateComment(c.Context(), &proto.UpdateCommentRequest{
		Comment: cmt,
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{
				`source_type`,
				`source`,
			},
		},
	})
	if err != nil {
		panic(err)
	}
}
