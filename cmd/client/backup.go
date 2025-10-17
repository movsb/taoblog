package client

import (
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/syncer"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/movsb/taorm"
	"github.com/spf13/cobra"
)

// BackupPosts backups all blog database.
func (c *Client) Backup(cmd *cobra.Command, compress bool, removeLogs bool) {
	backupClient, err := c.Management.Backup(c.Context(), &proto.BackupRequest{
		ClientDatabaseVersion: int32(migration.MaxVersionNumber()),
		Compress:              compress,
		RemoveLogs:            removeLogs,
	})
	if err != nil {
		log.Fatalln(err)
	}
	defer backupClient.CloseSend()

	bpr := &_BackupProgressReader{c: backupClient}

	var r io.ReadCloser
	if compress {
		zr, err := zlib.NewReader(bpr)
		if err != nil {
			panic(err)
		}
		r = zr
	} else {
		r = io.NopCloser(bpr)
	}
	defer r.Close()

	localDir := `.`
	name := `posts.db`
	localPath := filepath.Join(localDir, name)
	fp, err := os.Create(localPath)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	if _, err := io.Copy(fp, r); err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Printf("Filename: %s\n", localPath)

	c.BackupFiles()
}

type _BackupProgressReader struct {
	c         proto.Management_BackupClient
	d         []byte
	preparing bool
}

// Read ...
func (r *_BackupProgressReader) Read(p []byte) (int, error) {
	if len(r.d) == 0 {
		rsp, err := r.c.Recv()
		if err != nil {
			if err == io.EOF {
				return 0, err
			}
			log.Fatalln(err)
		}
		switch typed := rsp.BackupResponseMessage.(type) {
		case *proto.BackupResponse_Preparing_:
			fmt.Printf("\r\033[KPreparing... %.2f%%", typed.Preparing.Progress*100)
			r.preparing = true
		case *proto.BackupResponse_Transferring_:
			if r.preparing {
				fmt.Println()
				r.preparing = false
			}
			fmt.Printf("\r\033[KTransferring... %.2f%%", typed.Transferring.Progress*100)
			r.d = typed.Transferring.Data
		}
	}

	n := copy(p, r.d)
	r.d = r.d[n:]
	return n, nil
}

type SpecWithPostID struct {
	PostID int
	Path   string
	Digest string
}

func cmpSpecWithPostID(s, than SpecWithPostID) int {
	if s.PostID < than.PostID {
		return -1
	}
	if s.PostID > than.PostID {
		return +1
	}
	return strings.Compare(s.Digest, than.Digest)
}

func (s SpecWithPostID) Compare(than SpecWithPostID) int {
	return cmpSpecWithPostID(s, than)
}

func (s SpecWithPostID) DeepEqual(than SpecWithPostID) bool {
	return s.PostID == than.PostID && s.Digest == than.Digest
}

type Digest2Path struct {
	PostID int
	Digest string
}

func (c *Client) BackupFiles() {
	client, err := c.Management.BackupFiles(c.Context())
	if err != nil {
		log.Fatalln(err)
	}
	defer client.CloseSend()

	localDB := migration.InitFiles(`files.db`)
	postsDB := migration.InitPosts(`posts.db`, true, false)
	dataStore := storage.NewDataStore(taorm.NewDB(localDB))
	postsStore := storage.NewSQLite(taorm.NewDB(postsDB), dataStore, nil)

	var localSpecs, remoteSpecs []SpecWithPostID
	digest2path := make(map[Digest2Path]string)

	remoteFiles := utils.Must1(postsStore.AllFiles())
	for postID, files := range remoteFiles {
		for _, f := range files {
			remoteSpecs = append(remoteSpecs, SpecWithPostID{
				PostID: postID,
				Path:   f.Path,
				Digest: f.Digest,
			})
			digest2path[Digest2Path{PostID: postID, Digest: f.Digest}] = f.Path
		}
	}

	localFiles := utils.Must1(dataStore.ListAllFiles())
	for _, f := range localFiles {
		path := digest2path[Digest2Path{PostID: f.PostID, Digest: f.Digest}]
		if path == `` {
			path = fmt.Sprintf(`deleted:%s`, f.Digest)
		}

		localSpecs = append(localSpecs, SpecWithPostID{
			PostID: f.PostID,
			Path:   path,
			Digest: f.Digest,
		})
	}

	slices.SortFunc(remoteSpecs, cmpSpecWithPostID)
	slices.SortFunc(localSpecs, cmpSpecWithPostID)

	sync := syncer.New(
		syncer.WithCopyRemoteToLocal[[]SpecWithPostID](func(f SpecWithPostID) error {
			log.Println(`远程→本地：`, f.PostID, f.Path)
			err := client.Send(&proto.BackupFilesRequest{
				BackupFilesMessage: &proto.BackupFilesRequest_SendFile{
					SendFile: &proto.BackupFilesRequest_SendFileRequest{
						PostId: int32(f.PostID),
						Path:   f.Path,
					},
				},
			})
			if err != nil {
				panic(err)
			}
			rsp, err := client.Recv()
			if err != nil {
				panic(err)
			}
			if rsp.GetSendFile() == nil {
				panic(`bad message`)
			}
			data := rsp.GetSendFile().Data
			utils.Must(dataStore.CreateData(f.PostID, f.Digest, data))
			return nil
		}),
		syncer.WithDeleteLocal[[]SpecWithPostID](func(f SpecWithPostID) error {
			log.Println(`删除 本地：`, f.PostID, f.Path)
			utils.Must(dataStore.DeleteData(f.PostID, f.Digest))
			return nil
		}),
	)
	if err := sync.Sync(localSpecs, remoteSpecs, syncer.RemoteToLocal); err != nil {
		log.Fatalln(err)
	}
}
