package service

import (
	"bytes"
	"compress/zlib"
	"context"
	"database/sql"
	"io"
	"io/ioutil"
	"os"

	"github.com/mattn/go-sqlite3"
	"github.com/movsb/taoblog/protocols"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type _ReadCloseSizer struct {
	io.ReadCloser
	Size func() int
}

// Backup ...
func (s *Service) Backup(req *protocols.BackupRequest, srv protocols.Management_BackupServer) error {
	if !s.auth.AuthGRPC(srv.Context()).IsAdmin() {
		return status.Error(codes.Unauthenticated, "bad credentials")
	}

	sendPreparingProgress := func(progress float32) error {
		return srv.Send(&protocols.BackupResponse{
			BackupResponseMessage: &protocols.BackupResponse_Preparing_{
				Preparing: &protocols.BackupResponse_Preparing{
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
			zap.S().Errorw(`compress failed`, `err`, err)
			panic(`compress failed`)
		}
		if err := w.Close(); err != nil {
			zap.S().Errorw(`close failed`, `err`, err)
			panic(`close failed`)
		}
		zap.S().Infow(`compress completed`, `before`, n, `after`, buf.Len())
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
		return srv.Send(&protocols.BackupResponse{
			BackupResponseMessage: &protocols.BackupResponse_Transfering_{
				Transfering: &protocols.BackupResponse_Transfering{
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

	if err := dstConn.Raw(func(dstDC interface{}) error {
		rawDstConn := dstDC.(*sqlite3.SQLiteConn)

		srcConn, err := s.db.Conn(ctx)
		if err != nil {
			return err
		}
		defer srcConn.Close()

		return srcConn.Raw(func(srcDC interface{}) error {
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

	zap.L().Info(`backed up to file`, zap.String(`path`, tmpFile.Name()))

	return tmpFile.Name(), nil
}
