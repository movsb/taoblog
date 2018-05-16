package main

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"os/exec"
)

type BlogBackup struct {
	db *sql.DB
}

func NewBlogBackup(db *sql.DB) *BlogBackup {
	return &BlogBackup{
		db: db,
	}
}

func (o *BlogBackup) Backup(w io.Writer) error {
	opts := []string{
		"--add-drop-database",
		"--add-drop-table",
		"--add-locks",
		"--comments",
		"--compress",
	}

	opts = append(opts, "--user="+config.username)
	opts = append(opts, "--password="+config.password)
	opts = append(opts, "--databases", config.database)

	cmd := exec.Command(
		"mysqldump",
		opts...,
	)

	cmd.Stdout = w

	eb := bytes.NewBuffer(nil)
	cmd.Stderr = eb

	if cmd.Run() == nil {
		return nil
	}

	return errors.New(eb.String())
}
