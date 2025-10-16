package client

import (
	"io/fs"
	"log"
	"slices"

	client_common "github.com/movsb/taoblog/cmd/client/common"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
)

func (c *Client) DeletePost(id int64) error {
	_, err := c.Blog.DeletePost(c.Context(), &proto.DeletePostRequest{
		Id: int32(id),
	})
	return err
}

// UploadPostFiles 上传文章附件。
// TODO 应该像 Backup 那样改成带进度的 protocol buffer 方式上传。
// NOTE 路径列表，相对于工作目录，相对路径。
// TODO 由于评论中可能也带有图片引用，但是不会被算计到。所以远端的多余文件总是不会被删除。
// NOTE 会自动去重本地文件。
// NOTE 会自动排除 config.yml 文件。
func UploadPostFiles(client *clients.ProtoClient, id int64, root fs.FS, files []string) {
	files = slices.DeleteFunc(files, func(f string) bool { return f == client_common.ConfigFileName })

	if len(files) <= 0 {
		return
	}

	manage, err := client.Management.FileSystem(client.Context())
	if err != nil {
		panic(err)
	}
	defer manage.CloseSend()

	fsync := NewFilesSyncer(manage)

	if err := fsync.SyncPostFiles(id, root, files); err != nil {
		log.Fatalln(err)
	}
}
