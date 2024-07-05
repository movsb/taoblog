package daemon

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/spf13/cobra"
)

const updateScript = `docker compose pull taoblog && docker compose up -d taoblog`

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
			update()
		},
	}
	parent.AddCommand(daemonCmd)
}

func update() {
	// TODO 写死了。
	client := clients.NewFromGrpcAddr(`127.0.0.1:2563`)
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for range ticker.C {
		info, err := client.GetInfo(context.Background(), &proto.GetInfoRequest{})
		if err != nil {
			log.Println(err)
			continue
		}
		if !info.ScheduledUpdate {
			continue
		}
		log.Println(`需要更新。`)
		if err := run(updateScript); err != nil {
			log.Println(err)
		}
	}
}
