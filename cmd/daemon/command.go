package daemon

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/spf13/cobra"
)

const (
	updateScript = `docker compose pull taoblog && docker compose up -d taoblog`
)

func run(script string) error {
	cmd := exec.Command(`bash`, `-c`, script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func AddCommands(parent *cobra.Command) {
	daemonCmd := &cobra.Command{
		Use:   `daemon`,
		Short: `守护进程（更新镜像等）`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			home := utils.Must1(cmd.Flags().GetString(`home`))
			token := utils.Must1(cmd.Flags().GetString(`token`))
			update(home, token)
		},
	}
	daemonCmd.Flags().String(`home`, `http://127.0.0.1:2564`, `首页地址`)
	daemonCmd.Flags().StringP(`token`, `t`, ``, `预共享密钥`)
	parent.AddCommand(daemonCmd)
}

func update(home string, token string) {
	client := clients.NewFromHome(home, token)
	for {
		time.Sleep(time.Second * 15)
		info, err := client.Blog.GetInfo(context.Background(), &proto.GetInfoRequest{})
		if err != nil {
			log.Println(`GetInfo:`, err)
			continue
		}
		if !info.ScheduledUpdate {
			continue
		}
		log.Println(`检测到设置了更新标识，需要更新。`)
		if err := run(updateScript); err != nil {
			log.Println(err)
			continue
		}
	}
}
