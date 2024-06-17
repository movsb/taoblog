package webhook

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func createReloader(script string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		err := run(script)
		if err == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
}

func run(script string) error {
	cmd := exec.Command(`bash`, `-c`, script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func AddCommands(parent *cobra.Command) {
	var (
		listen      string
		script      string
		immediately bool
	)

	reloadCmd := &cobra.Command{
		Use:   `reload`,
		Short: `GitHub Webhook Reloader`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if immediately {
				if err := run(script); err != nil {
					log.Fatalln(err)
				}
				return
			}
			http.HandleFunc(`/reload`, createReloader(script))
			l, err := net.Listen(`unix`, listen)
			if err != nil {
				log.Fatalln(err)
			}
			if err := http.Serve(l, nil); err != nil {
				log.Fatalln(err)
			}
		},
	}
	reloadCmd.Flags().BoolVarP(&immediately, `immediately`, `r`, false, `立即执行。`)
	reloadCmd.Flags().StringVarP(&listen, `listen`, `l`, `/tmp/taoblog-reloader.sock`, `监控 /reload 等待执行。`)
	reloadCmd.Flags().StringVarP(&script, `script`, `s`, `docker compose pull taoblog && docker compose up -d taoblog`, `拉取新镜像并重启 TaoBlog。`)
	parent.AddCommand(reloadCmd)
}
