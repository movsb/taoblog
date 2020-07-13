package service

import (
	"bytes"
	"compress/zlib"
	"context"

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

	if s.cfg.Database.Engine != `sqlite` {
		panic(`sqlite only`)
	}

	panic(`not supported`)

	ob := bytes.NewBuffer(nil)

	before := ob.Bytes()

	if req.Compress {
		buf := bytes.NewBuffer(nil)
		w, err := zlib.NewWriterLevel(buf, zlib.BestCompression)
		if err != nil {
			panic(err)
		}
		if n, err := w.Write(before); n != len(before) || err != nil {
			zap.S().Errorw(`compress failed`, `n`, n, `len`, len(before), `err`, err)
			panic(`compress failed`)
		}
		if err := w.Close(); err != nil {
			zap.S().Errorw(`close failed`, `err`, err)
			panic(`close failed`)
		}
		zap.S().Infow(`compress completed`, `before`, len(before), `after`, buf.Len())
		before = buf.Bytes()
	}

	return &protocols.BackupResponse{
		Data: before,
	}, nil
}
