package tools

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/spf13/cobra"
)

func AddCommands(parent *cobra.Command) {
	serveCmd := cobra.Command{
		Use:   `serve`,
		Short: `运行一个临时的 HTTP 服务器。`,
		Run: func(cmd *cobra.Command, args []string) {
			dir := utils.Must1(cmd.Flags().GetString(`dir`))
			port := utils.Must1(cmd.Flags().GetInt(`port`))
			fs := http.FileServerFS(os.DirFS(dir))
			if err := http.ListenAndServe(fmt.Sprintf(`:%d`, port), fs); err != http.ErrServerClosed {
				log.Fatalln(err)
			}
		},
	}
	serveCmd.Flags().StringP(`dir`, `d`, `.`, `目录。`)
	serveCmd.Flags().IntP(`port`, `p`, 9090, `端口。`)

	parent.AddCommand(&serveCmd)
}
