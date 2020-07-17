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

// Backup ...
func (s *Service) Backup(ctx context.Context, req *protocols.BackupRequest) (*protocols.BackupResponse, error) {
	if !s.auth.AuthGRPC(ctx).IsAdmin() {
		return nil, status.Error(codes.Unauthenticated, "bad credentials")
	}

	var r io.ReadCloser

	switch s.cfg.Database.Engine {
	default:
		panic(`engine not supported`)
	case `sqlite`:
		path, err := s.backupSQLite3(ctx)
		if err != nil {
			return nil, err
		}
		defer os.Remove(path)
		fp, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		r = fp
	}

	defer r.Close()

	if req.Compress {
		buf := bytes.NewBuffer(nil)
		w, err := zlib.NewWriterLevel(buf, zlib.BestCompression)
		if err != nil {
			panic(err)
		}
		n, err := io.Copy(w, r)
		if err != nil {
			zap.S().Errorw(`compress failed`, `err`, err)
			panic(`compress failed`)
		}
		if err := w.Close(); err != nil {
			zap.S().Errorw(`close failed`, `err`, err)
			panic(`close failed`)
		}
		zap.S().Infow(`compress completed`, `before`, n, `after`, buf.Len())
		r = ioutil.NopCloser(buf)
	}

	all, err := ioutil.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return &protocols.BackupResponse{
		Data: all,
	}, nil
}

func (s *Service) backupSQLite3(ctx context.Context) (string, error) {
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

		if err := srcConn.Raw(func(srcDC interface{}) error {
			rawSrcConn := srcDC.(*sqlite3.SQLiteConn)
			backup, err := rawDstConn.Backup(`main`, rawSrcConn, `main`)
			if err != nil {
				return err
			}

			// errors can be safely ignored.
			_, _ = backup.Step(-1)

			if err := backup.Close(); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return ``, err
	}

	zap.L().Info(`backuped to file`, zap.String(`path`, tmpFile.Name()))

	return tmpFile.Name(), nil
}
