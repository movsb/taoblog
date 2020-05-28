package service

import (
	"bytes"
	"context"
	"net"
	"os/exec"

	"github.com/movsb/taoblog/protocols"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Backup ...
func (s *Service) Backup(ctx context.Context, req *protocols.BackupRequest) (*protocols.BackupResponse, error) {
	if !s.auth.AuthGRPC(ctx).IsAdmin() {
		return nil, status.Error(codes.Unauthenticated, "bad credentials")
	}

	if s.cfg.Database.Engine != `mysql` {
		panic(`mysql only`)
	}

	opts := []string{
		"--add-drop-database",
		"--add-drop-table",
		"--add-locks",
		"--comments",
		"--compress",
	}

	host, port, _ := net.SplitHostPort(s.cfg.Database.MySQL.Endpoint)

	opts = append(opts, "--host="+host)
	opts = append(opts, "--port="+port)
	opts = append(opts, "--user="+s.cfg.Database.MySQL.Username)
	opts = append(opts, "--password="+s.cfg.Database.MySQL.Password)
	opts = append(opts, "--databases", s.cfg.Database.MySQL.Database)

	cmd := exec.Command(
		"mysqldump",
		opts...,
	)

	ob := bytes.NewBuffer(nil)
	cmd.Stdout = ob
	eb := bytes.NewBuffer(nil)
	cmd.Stderr = eb

	if err := cmd.Run(); err != nil {
		return nil, status.Errorf(codes.Internal, `%v:%s`, err, eb.String())
	}

	return &protocols.BackupResponse{
		Data: ob.String(),
	}, nil
}
