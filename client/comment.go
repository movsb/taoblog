package client

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/movsb/taoblog/protocols"
	field_mask "google.golang.org/protobuf/types/known/fieldmaskpb"
)

// SetCommentPostID ...
func (c *Client) SetCommentPostID(commentID int64, postID int64) {
	req := protocols.SetCommentPostIDRequest{
		Id:     commentID,
		PostId: postID,
	}
	resp, err := c.blog.SetCommentPostID(c.token(), &req)
	if err != nil {
		panic(err)
	}
	_ = resp
}

func (c *Client) GetComment(cmdID int64) *protocols.Comment {
	cmt, err := c.blog.GetComment(c.token(), &protocols.GetCommentRequest{
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

	source := cmt.Source
	if source == `` || cmt.SourceType != `markdown` {
		source = cmt.Content
	}

	if _, err := tmpFile.WriteString(source); err != nil {
		panic(err)
	}

	oldInfo, err := tmpFile.Stat()
	if err != nil {
		panic(err)
	}

	tmpFile.Close()

	fmt.Printf("Editing comment: %d, post: %d\n", cmt.Id, cmt.PostId)

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(err)
	}

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

	cmt.SourceType = `markdown`
	cmt.Source = string(bys)

	_, err = c.blog.UpdateComment(c.token(), &protocols.UpdateCommentRequest{
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
