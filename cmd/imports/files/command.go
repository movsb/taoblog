package files

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/movsb/taoblog/cmd/server"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/spf13/cobra"
)

func CreateCommands() *cobra.Command {
	filesCmd := &cobra.Command{
		Use:   `files <name.db> <dir>`,
		Short: `把文章附件目录导入成数据库文件`,
		Run:   importFiles,
		Args:  cobra.ExactArgs(2),
	}

	return filesCmd
}

func importFiles(cmd *cobra.Command, args []string) {
	name, dir := args[0], args[1]
	db := server.InitDatabase(name, server.InitForFiles())
	files := utils.Must1(utils.ListFiles(os.DirFS(dir), "."))
	dbfs := storage.NewSQLite(db)
	defer db.Close()

	for _, f := range files {
		func() {
			log.Println(`导入：`, f.Path)

			r := utils.Must1(os.DirFS(dir).Open(f.Path))
			defer r.Close()

			dir, path, _ := strings.Cut(f.Path, `/`)
			pid := utils.Must1(strconv.Atoi(dir))
			f.Path = path

			fs := utils.Must1(dbfs.ForPost(pid))
			utils.Must(utils.Write(fs, f, r))
		}()
	}
}
