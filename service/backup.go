package service

import (
	"bytes"
	"errors"
	"os"
	"os/exec"

	"github.com/movsb/taoblog/protocols"
)

func (s *ImplServer) GetBackup(in *protocols.GetBackupRequest) *protocols.Empty {
	opts := []string{
		"--add-drop-database",
		"--add-drop-table",
		"--add-locks",
		"--comments",
		"--compress",
	}

	opts = append(opts, "--user="+os.Getenv("DB_USERNAME"))
	opts = append(opts, "--password="+os.Getenv("DB_PASSWORD"))
	opts = append(opts, "--databases", os.Getenv("DB_DATABASE"))

	cmd := exec.Command(
		"mysqldump",
		opts...,
	)

	cmd.Stdout = in.W

	eb := bytes.NewBuffer(nil)
	cmd.Stderr = eb

	if cmd.Run() == nil {
		return nil
	}

	panic(errors.New(eb.String()))
}
