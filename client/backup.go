package client

import (
	"compress/zlib"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"github.com/spf13/cobra"
)

// BackupPosts backups all blog database.
func (c *Client) BackupPosts(cmd *cobra.Command) {
	compress := true

	backupClient, err := c.management.Backup(c.token(), &protocols.BackupRequest{Compress: compress})
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
		r = ioutil.NopCloser(bpr)
	}
	defer r.Close()

	var w io.Writer
	bStdout, err := cmd.Flags().GetBool(`stdout`)
	if err != nil {
		panic(err)
	}
	if !bStdout {
		name := time.Now().Format(`taoblog.2006-01-02.db`)
		fp, err := os.Create(name)
		if err != nil {
			panic(err)
		}
		defer fmt.Printf("Filename: %s\n", name)
		defer fp.Close()
		w = fp

		bNoLink, err := cmd.Flags().GetBool((`no-link`))
		if err != nil {
			panic(err)
		}
		if !bNoLink {
			defer func() {
				link := `taoblog.db`
				if _, err := os.Stat(link); err == nil {
					os.Remove(link)
				}
				err := os.Symlink(name, link)
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
	c         protocols.Management_BackupClient
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
			panic(err)
		}
		switch typed := rsp.BackupResponseMessage.(type) {
		case *protocols.BackupResponse_Preparing_:
			fmt.Printf("\r\033[KPreparing... %.2f%%", typed.Preparing.Progress*100)
			r.preparing = true
		case *protocols.BackupResponse_Transfering_:
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

// Backup backups all blog database.
func (c *Client) BackupFiles(cmd *cobra.Command) {
	client, err := c.management.BackupFiles(c.token())
	if err != nil {
		panic(err)
	}
	defer client.CloseSend()
	log.Printf(`Remote: list files...`)
	err = client.Send(&protocols.BackupFilesRequest{
		BackupFilesMessage: &protocols.BackupFilesRequest_ListFiles{
			ListFiles: &protocols.BackupFilesRequest_ListFilesRequest{},
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

	log.Printf(`Local: list files...`)
	localDir := `./files`
	localFiles, err := utils.ListBackupFiles(localDir)
	if err != nil {
		panic(err)
	}

	log.Println(`Sort files...`)
	sort.Slice(remoteFiles, func(i, j int) bool {
		return strings.Compare(remoteFiles[i].Path, remoteFiles[j].Path) < 0
	})
	sort.Slice(localFiles, func(i, j int) bool {
		return strings.Compare(localFiles[i].Path, localFiles[j].Path) < 0
	})

	// yaml.NewEncoder(os.Stdout).Encode(remoteFiles)
	// yaml.NewEncoder(os.Stdout).Encode(localFiles)

	deleteLocal := func(f *protocols.BackupFileSpec) {
		path := filepath.Join(localDir, f.Path)
		if err := os.Remove(path); err != nil {
			panic(err)
		}
	}
	copyRemote := func(f *protocols.BackupFileSpec) {
		localPath := filepath.Join(localDir, f.Path)
		mode := os.FileMode(f.Mode)
		if mode.IsDir() {
			if err := os.MkdirAll(localPath, mode.Perm()); err != nil {
				panic(err)
			}
			return
		}
		dir := filepath.Dir(localPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(err)
		}
		err := client.Send(&protocols.BackupFilesRequest{
			BackupFilesMessage: &protocols.BackupFilesRequest_SendFile{
				SendFile: &protocols.BackupFilesRequest_SendFileRequest{
					Path: f.Path,
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
		fp, err := os.OpenFile(localPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())
		if err != nil {
			panic(err)
		}
		defer func(fp *os.File) {
			if err := fp.Close(); err != nil {
				panic(err)
			}
		}(fp)
		if _, err := fp.Write(data); err != nil {
			panic(err)
		}
		t := time.Unix(int64(f.Time), 0)
		if err := os.Chtimes(localPath, t, t); err != nil {
			panic(err)
		}
	}

	rf, lf := remoteFiles, localFiles
	i, j := len(rf)-1, len(lf)-1

	for {
		if i == -1 && j == -1 {
			log.Println(`No more files to compare`)
			break
		}
		if j == -1 {
			log.Printf("Local: copy %s\n", rf[i].Path)
			copyRemote(rf[i])
			i--
			continue
		}
		if i == -1 {
			log.Printf("Local: delete %s\n", lf[j].Path)
			deleteLocal(lf[j])
			j--
			continue
		}
		switch n := strings.Compare(rf[i].Path, lf[j].Path); {
		case n < 0:
			log.Printf("Local: delete %s\n", lf[j].Path)
			deleteLocal(lf[j])
			j--
		case n == 0:
			rm, lm := os.FileMode(rf[i].Mode), os.FileMode(lf[j].Mode)
			if rm.IsDir() != lm.IsDir() {
				panic(`file != dir`)
			}
			shouldSync := false
			if rf[i].Size != lf[j].Size {
				shouldSync = true
			}
			if rf[i].Time != lf[j].Time {
				shouldSync = true
			}
			if shouldSync {
				if rm.IsRegular() {
					log.Printf("Local: delete %s\n", lf[j].Path)
					deleteLocal(lf[j])
					log.Printf("Local: copy %s\n", rf[i].Path)
					copyRemote(rf[i])
				} else {
					log.Printf("Local: modtime of dir: %s", rf[i].Path)
					path := filepath.Join(localDir, lf[j].Path)
					t := time.Unix(int64(rf[i].Time), 0)
					if err := os.Chtimes(path, t, t); err != nil {
						panic(err)
					}
				}
			}
			i--
			j--
		case n > 0:
			log.Printf("Local: copy %s\n", rf[i].Path)
			copyRemote(rf[i])
			i--
		}
	}
}
