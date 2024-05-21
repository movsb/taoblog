package service

import (
	"bytes"
	"compress/zlib"
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/mattn/go-sqlite3"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type _ReadCloseSizer struct {
	io.ReadCloser
	Size func() int
}

// Backup ...
func (s *Service) Backup(req *proto.BackupRequest, srv proto.Management_BackupServer) error {
	s.MustBeAdmin(srv.Context())

	sendPreparingProgress := func(progress float32) error {
		return srv.Send(&proto.BackupResponse{
			BackupResponseMessage: &proto.BackupResponse_Preparing_{
				Preparing: &proto.BackupResponse_Preparing{
					Progress: progress,
				},
			},
		})
	}

	var rcs _ReadCloseSizer

	switch s.cfg.Database.Engine {
	default:
		panic(`engine not supported`)
	case `sqlite`:
		path, err := s.backupSQLite3(srv.Context(), sendPreparingProgress)
		if err != nil {
			return err
		}
		defer os.Remove(path)
		fp, err := os.Open(path)
		if err != nil {
			return err
		}
		rcs.ReadCloser = fp
		rcs.Size = func() int {
			stat, _ := fp.Stat()
			return int(stat.Size())
		}
	}

	defer rcs.Close()

	if req.Compress {
		buf := bytes.NewBuffer(nil)
		w, err := zlib.NewWriterLevel(buf, zlib.BestCompression)
		if err != nil {
			panic(err)
		}
		n, err := io.Copy(w, rcs)
		if err != nil {
			log.Println(`compress failed:`, err)
			panic(`compress failed`)
		}
		if err := w.Close(); err != nil {
			log.Println(`close failed:`, err)
			panic(`close failed`)
		}
		log.Printf(`compress completed: before: %v, after: %v`, n, buf.Len())
		rcs.ReadCloser = ioutil.NopCloser(buf)
		rcs.Size = func() int {
			return buf.Len()
		}
	}

	const (
		minSize = 16 << 10
		maxSize = 1 << 20
	)
	totalSize := rcs.Size()
	stepSize := totalSize / 100
	switch {
	case stepSize < minSize:
		stepSize = minSize
	case stepSize > maxSize:
		stepSize = maxSize
	}

	sendTransfer := func(data []byte, progress float32) error {
		return srv.Send(&proto.BackupResponse{
			BackupResponseMessage: &proto.BackupResponse_Transfering_{
				Transfering: &proto.BackupResponse_Transfering{
					Progress: progress,
					Data:     data,
				},
			},
		})
	}

	buf := make([]byte, stepSize)
	sent := 0
	for {
		n, err := rcs.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		sent += n
		if err = sendTransfer(buf[:n], float32(sent)/float32(totalSize)); err != nil {
			return err
		}
	}

	return nil
}

// https://www.sqlite.org/c3ref/backup_finish.html
func (s *Service) backupSQLite3(ctx context.Context, progress func(percentage float32) error) (string, error) {
	tmpFile, err := ioutil.TempFile(``, `taoblog-*`)
	if err != nil {
		return ``, err
	}
	tmpFile.Close()

	dstDB, err := sql.Open(`sqlite3`, tmpFile.Name())
	if err != nil {
		return ``, err
	}
	defer dstDB.Close()

	dstConn, err := dstDB.Conn(ctx)
	if err != nil {
		return ``, err
	}
	defer dstConn.Close()

	if err := dstConn.Raw(func(dstDC any) error {
		rawDstConn := dstDC.(*sqlite3.SQLiteConn)

		srcConn, err := s.db.Conn(ctx)
		if err != nil {
			return err
		}
		defer srcConn.Close()

		return srcConn.Raw(func(srcDC any) error {
			rawSrcConn := srcDC.(*sqlite3.SQLiteConn)
			backup, err := rawDstConn.Backup(`main`, rawSrcConn, `main`)
			if err != nil {
				return err
			}
			defer backup.Close() // close twice

			if progress == nil {
				progress = func(p float32) error {
					// fmt.Println(p)
					return nil
				}
			}

			var (
				remaining int
				total     int
				step      int
			)

			for {
				done, err := backup.Step(step)
				if err != nil {
					return err
				}

				// will keep update by sqlite3_backup_step()
				remaining = backup.Remaining()
				total = backup.PageCount()
				if step == 0 {
					step = total / 10
					if step < 1 {
						step = 10
					}
				}

				if err := progress(1 - float32(remaining)/float32(total)); err != nil {
					return err
				}

				if done {
					break
				}
			}

			return backup.Close()
		})
	}); err != nil {
		return ``, err
	}

	log.Printf(`backed up to file: path: %s`, tmpFile.Name())

	return tmpFile.Name(), nil
}

func (s *Service) BackupFiles(srv proto.Management_BackupFilesServer) error {
	s.MustBeAdmin(srv.Context())

	listFiles := func(req *proto.BackupFilesRequest_ListFilesRequest) error {
		files, err := utils.ListBackupFiles(s.cfg.Data.File.Path)
		if err != nil {
			log.Printf(`BackupFiles failed to list files: %v`, err)
			return err
		}
		rsp := &proto.BackupFilesResponse{
			BackupFilesMessage: &proto.BackupFilesResponse_ListFiles{
				ListFiles: &proto.BackupFilesResponse_ListFilesResponse{
					Files: files,
				},
			},
		}
		if err := srv.Send(rsp); err != nil {
			log.Printf(`BackupFiles failed to send file list: %v`, err)
			return err
		}
		return nil
	}

	sendFile := func(req *proto.BackupFilesRequest_SendFileRequest) error {
		log.Printf("send file: %s", req.Path)
		localPath := filepath.Join(s.cfg.Data.File.Path, filepath.Clean(req.Path))
		data, err := ioutil.ReadFile(localPath)
		if err != nil {
			return err
		}
		rsp := &proto.BackupFilesResponse{
			BackupFilesMessage: &proto.BackupFilesResponse_SendFile{
				SendFile: &proto.BackupFilesResponse_SendFileResponse{
					Data: data,
				},
			},
		}
		if err := srv.Send(rsp); err != nil {
			log.Printf(`BackupFiles failed to send file: %v`, err)
			return err
		}
		return nil
	}

	for {
		req, err := srv.Recv()
		if err != nil {
			if err == io.EOF {
				log.Printf(`BackupFiles finished`)
				return nil
			}
			if st, ok := status.FromError(err); ok && st.Code() == codes.Canceled {
				log.Printf(`BackupFiles finished`)
				return nil
			}
			log.Printf(`BackupFiles failed: %v`, err)
			return err
		}
		if req := req.GetListFiles(); req != nil {
			if err := listFiles(req); err != nil {
				return err
			}
			continue
		}
		if req := req.GetSendFile(); req != nil {
			if err := sendFile(req); err != nil {
				return err
			}
			continue
		}
		return fmt.Errorf(`unknown message`)
	}
}
