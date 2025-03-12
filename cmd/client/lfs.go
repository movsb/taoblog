package client

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	client_common "github.com/movsb/taoblog/cmd/client/common"
	"github.com/spf13/cobra"
)

const lfsName = `.lfs`

func createLfsCommands() *cobra.Command {
	// GIT 仓库根目录
	var root string

	lfsCmd := &cobra.Command{
		Short: `大文件操作命令。`,
		Use:   `lfs`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			buf := bytes.NewBuffer(nil)
			if err := spawnWithOutput(`git`, []string{`rev-parse`, `--show-toplevel`}, ".", "", buf); err != nil {
				log.Fatalln(err)
			}
			dir := strings.TrimSpace(buf.String())
			root = dir
		},
	}
	lfsAddCmd := &cobra.Command{
		Short: `添加文件到大文件仓库。`,
		Use:   `add <files...>`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config, err := client_common.ReadPostConfig(client_common.ConfigFileName)
			if err != nil {
				log.Fatalln(err)
			}
			if config.ID < 0 {
				log.Fatalln(`无效文章编号。`)
			}
			// TODO 提供的文件可能包含目录。
			dir := filepath.Join(root, lfsName, fmt.Sprint(config.ID))
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalln(err, dir)
			}
			for _, f := range args {
				cleaned := filepath.Clean(f)
				if filepath.IsAbs(cleaned) {
					log.Fatalln(`错误的路径。`)
				}

				stat, err := os.Lstat(f)
				if err != nil {
					log.Fatalln(err)
				}
				if stat.IsDir() {
					log.Fatalln(`不支持目录。`)
				}
				if stat.Mode().Type() == fs.ModeSymlink {
					log.Println(`忽略软连接文件。`)
					continue
				}

				if err := os.Rename(cleaned, dir); err != nil {
					log.Fatalln(err)
				}

				if err := os.Symlink(filepath.Join(dir, cleaned), filepath.Base(cleaned)); err != nil {
					log.Fatalln(err)
				}
			}
		},
	}
	lfsCmd.AddCommand(lfsAddCmd)
	lfsCommitCmd := &cobra.Command{
		Short: `提交保存。`,
		Use:   `commit`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			config, err := client_common.ReadPostConfig(client_common.ConfigFileName)
			if err != nil {
				log.Fatalln(err)
			}
			if config.ID < 0 {
				log.Fatalln(`无效文章编号。`)
			}
			// TODO 提供的文件可能包含目录。
			dir := filepath.Join(lfsName, fmt.Sprint(config.ID))
			if err := spawn(`git`, []string{`add`, `.`}, dir, ""); err != nil {
				log.Fatalln(err)
			}
			if err := spawn(`git`, []string{`commit`, `-m`, `by lfs cmd`}, dir, ""); err != nil {
				log.Fatalln(err)
			}
			if err := spawn(`git`, []string{`add`}, filepath.Join(root, lfsName), ""); err != nil {
				log.Fatalln(err)
			}
		},
	}
	lfsCmd.AddCommand(lfsCommitCmd)
	return lfsCmd
}

func spawn(name string, args []string, dir string, input string) error {
	return spawnWithOutput(name, args, dir, input, os.Stdout)
}

func spawnWithOutput(name string, args []string, dir string, input string, output io.Writer) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = output
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
