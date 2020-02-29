package service

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

func (s *Service) GetBackup(w io.Writer) {
	opts := []string{
		"--add-drop-database",
		"--add-drop-table",
		"--add-locks",
		"--comments",
		"--compress",
	}

	opts = append(opts, "--user="+s.cfg.Database.Username)
	opts = append(opts, "--password="+s.cfg.Database.Password)
	opts = append(opts, "--databases", s.cfg.Database.Database)

	cmd := exec.Command(
		"mysqldump",
		opts...,
	)

	cmd.Stdout = w

	eb := bytes.NewBuffer(nil)
	cmd.Stderr = eb

	if cmd.Run() == nil {
		return
	}

	panic(errors.New(eb.String()))
}
