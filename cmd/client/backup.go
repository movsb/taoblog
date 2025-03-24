package client

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/utils/syncer"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taoblog/setup/migration"
	"github.com/spf13/cobra"
)

// BackupPosts backups all blog database.
func (c *Client) BackupPosts(cmd *cobra.Command) {
	compress := true

	backupClient, err := c.Management.Backup(c.Context(), &proto.BackupRequest{Compress: compress})
	if err != nil {
		panic(err)
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

	var w io.Writer
	bStdout, err := cmd.Flags().GetBool(`stdout`)
	if err != nil {
		panic(err)
	}
	if !bStdout {
		localDir := `./posts`
		if err := os.MkdirAll(localDir, 0755); err != nil {
			panic(err)
		}
		name := time.Now().Format(`taoblog.2006-01-02.db`)
		localPath := filepath.Join(localDir, name)
		fp, err := os.Create(localPath)
		if err != nil {
			panic(err)
		}
		defer fmt.Printf("Filename: %s\n", localPath)
		defer fp.Close()
		w = fp

		bNoLink, err := cmd.Flags().GetBool((`no-link`))
		if err != nil {
			panic(err)
		}
		if !bNoLink {
			defer func() {
				link := `posts.db`
				if _, err := os.Stat(link); err == nil {
					os.Remove(link)
				}
				err := os.Symlink(localPath, link)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
				fmt.Printf("Symlinked to %s\n", link)
			}()
		}
	} else {
		w = os.Stdout
	}

	if _, err := io.Copy(w, r); err != nil {
		panic(err)
	}

	fmt.Println()
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
		case *proto.BackupResponse_Transfering_:
			if r.preparing {
				fmt.Println()
				r.preparing = false
			}
			fmt.Printf("\r\033[KTransfering... %.2f%%", typed.Transfering.Progress*100)
			r.d = typed.Transfering.Data
		}
	}

	n := copy(p, r.d)
	r.d = r.d[n:]
	return n, nil
}

type SpecWithPostID struct {
	PostID int
	File   *proto.FileSpec
}

func lessSpecWithPostID(s, than SpecWithPostID) bool {
	return s.PostID < than.PostID || s.PostID == than.PostID && s.File.Path < than.File.Path
}

func (s SpecWithPostID) Less(than SpecWithPostID) bool {
	return lessSpecWithPostID(s, than)
}

func (s SpecWithPostID) DeepEqual(to SpecWithPostID) bool {
	return s.PostID == to.PostID &&
		s.File.Path == to.File.Path &&
		s.File.Time == to.File.Time &&
		s.File.Size == to.File.Size
}

func (c *Client) BackupFiles(cmd *cobra.Command, name string) {
	client, err := c.Management.BackupFiles(c.Context())
	if err != nil {
		log.Fatalln(err)
	}
	defer client.CloseSend()
	log.Printf(`Remote: list files...`)
	err = client.Send(&proto.BackupFilesRequest{
		BackupFilesMessage: &proto.BackupFilesRequest_ListFiles{
			ListFiles: &proto.BackupFilesRequest_ListFilesRequest{},
		},
	})
	if err != nil {
		panic(err)
	}
	rsp, err := client.Recv()
	if err != nil {
		panic(err)
	}
	if rsp.GetListFiles() == nil {
		panic(`bad message`)
	}
	remoteFiles := rsp.GetListFiles().Files

	var localSpecs, remoteSpecs []SpecWithPostID

	for id, r := range remoteFiles {
		for _, f := range r.Files {
			remoteSpecs = append(remoteSpecs, SpecWithPostID{
				PostID: int(id),
				File:   f,
			})
		}
	}

	log.Printf(`Local: list files...`)

	localDB := migration.InitFiles(name)
	localStore := storage.NewSQLite(localDB)
	localFiles, err := localStore.AllFiles()
	if err != nil {
		panic(err)
	}

	for id, r := range localFiles {
		for _, f := range r {
			localSpecs = append(localSpecs, SpecWithPostID{
				PostID: int(id),
				File:   f,
			})
		}
	}

	log.Println(`Sort files...`)
	sort.Slice(remoteSpecs, func(i, j int) bool {
		return remoteSpecs[i].PostID < remoteSpecs[j].PostID || strings.Compare(remoteSpecs[i].File.Path, remoteSpecs[j].File.Path) < 0
	})
	sort.Slice(localSpecs, func(i, j int) bool {
		return localSpecs[i].PostID < localSpecs[j].PostID || strings.Compare(localSpecs[i].File.Path, localSpecs[j].File.Path) < 0
	})

	sync := syncer.New(
		syncer.WithCopyRemoteToLocal[[]SpecWithPostID](func(f SpecWithPostID) error {
			log.Println(`远程→本地：`, f.PostID, f.File.Path)
			err := client.Send(&proto.BackupFilesRequest{
				BackupFilesMessage: &proto.BackupFilesRequest_SendFile{
					SendFile: &proto.BackupFilesRequest_SendFileRequest{
						PostId: int32(f.PostID),
						Path:   f.File.Path,
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
			fs := utils.Must1(localStore.ForPost(f.PostID))
			utils.Must(utils.Write(fs, f.File, bytes.NewReader(data)))
			return nil
		}),
		syncer.WithDeleteLocal[[]SpecWithPostID](func(f SpecWithPostID) error {
			log.Println(`删除本地：`, f.PostID, f.File.Path)
			fs := utils.Must1(localStore.ForPost(f.PostID))
			utils.Must(utils.Delete(fs, f.File.Path))
			return nil
		}),
	)
	if err := sync.Sync(localSpecs, remoteSpecs, syncer.RemoteToLocal); err != nil {
		log.Fatalln(err)
	}
}
