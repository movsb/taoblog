package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/movsb/taoblog/client/exec"
	"github.com/movsb/taoblog/protocols"
	"google.golang.org/genproto/protobuf/field_mask"
)

// SetCommentPostID ...
func (c *Client) SetCommentPostID(commentID int64, postID int64) {
	cmt := protocols.Comment{
		Id:     commentID,
		PostId: postID,
	}
	bys, err := json.Marshal(cmt)
	if err != nil {
		panic(err)
	}
	resp := c.mustPost(fmt.Sprintf(`/comments!setPostID`), bytes.NewReader(bys), contentTypeJSON)
	defer resp.Body.Close()
}

func (c *Client) GetComment(cmdID int64) *protocols.Comment {
	cmt, err := c.grpcClient.GetComment(c.token(), &protocols.GetCommentRequest{
		Id: cmdID,
	})
	if err != nil {
		panic(err)
	}
	return cmt
}

// UpdateComment ...
func (c *Client) UpdateComment(cmdID int64) {
	cmt := c.GetComment(cmdID)
	editor, ok := os.LookupEnv(`EDITOR`)
	if !ok {
		editor = `vim`
	}

	tmpFile, err := ioutil.TempFile(``, `taoblog-comment-`)
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(cmt.Content); err != nil {
		panic(err)
	}

	oldInfo, err := tmpFile.Stat()
	if err != nil {
		panic(err)
	}

	tmpFile.Close()

	exec.Exec(editor, tmpFile.Name())

	newInfo, err := os.Stat(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	if newInfo.ModTime() == oldInfo.ModTime() {
		fmt.Println(`file not modified`)
		return
	}

	bys, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		panic(err)
	}
	cmt.Content = string(bys)
	_, err = c.grpcClient.UpdateComment(c.token(), &protocols.UpdateCommentRequest{
		Comment: cmt,
		UpdateMask: &field_mask.FieldMask{
			Paths: []string{
				`content`,
			},
		},
	})
	if err != nil {
		panic(err)
	}
}
