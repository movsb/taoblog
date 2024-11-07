package sync

import (
	"log"
	"os"
	"time"

	"github.com/movsb/pkg/notify"
	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/spf13/cobra"
)

func AddCommands(parent *cobra.Command) {
	syncCmd := &cobra.Command{
		Use: `sync`,
		Run: func(cmd *cobra.Command, args []string) {
			chanifyToken := os.Getenv(`CHANIFY`)
			if chanifyToken == "" {
				panic(`empty CHANIFY`)
			}
			ch := notify.NewOfficialChanify(chanifyToken)

			full := utils.Must1(cmd.Flags().GetBool(`full`))

			gs := New(client.InitHostConfigs(), ".", full)
			for {
				if err := gs.Sync(); err != nil {
					ch.Send("同步失败", err.Error(), true)
					log.Println(err)
					time.Sleep(time.Minute * 15)
					continue
				} else {
					log.Println(`同步完成。`)
					ch.Send(`同步成功`, `全部完成，没有错误。`, false)
				}
				time.Sleep(time.Hour)
			}
		},
	}
	syncCmd.Flags().Bool(`full`, false, `初次备份是否全量扫描更新。`)
	parent.AddCommand(syncCmd)
}
