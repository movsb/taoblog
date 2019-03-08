package service

import (
	"bytes"
	"errors"
	"io"
	"os"
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

	opts = append(opts, "--user="+os.Getenv("DB_USERNAME"))
	opts = append(opts, "--password="+os.Getenv("DB_PASSWORD"))
	opts = append(opts, "--databases", os.Getenv("DB_DATABASE"))

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
