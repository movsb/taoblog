package sync

import (
	"log"
	"os"
	"time"

	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/spf13/cobra"
)

func AddCommands(parent *cobra.Command) {
	syncCmd := &cobra.Command{
		Use: `sync`,
		Run: func(cmd *cobra.Command, args []string) {
			full := utils.Must1(cmd.Flags().GetBool(`full`))
			every := utils.Must1(cmd.Flags().GetDuration(`every`))

			// TODO 直接用 API 从服务器拉取此配置。
			cred := Credential{
				Author:   os.Getenv(`AUTHOR`),
				Email:    os.Getenv(`EMAIL`),
				Username: os.Getenv(`USERNAME`),
				Password: os.Getenv(`PASSWORD`),
			}
			if cred.Author == `` || cred.Email == `` || cred.Username == `` || cred.Password == `` {
				log.Fatalln(`凭证为空。`)
			}

			config := client.InitHostConfigs()
			client := clients.NewProtoClient(config.Home, config.Token)

			gs := New(client, cred, ".", full)

			sync := func() error {
				if err := gs.Sync(); err != nil {
					client.SendInstant("同步失败", err.Error())
					return err
				}

				client.SendInstant(`同步成功`, `全部完成，没有错误。`)
				log.Println(`同步完成。`)
				return nil
			}

			if every <= 0 {
				if err := sync(); err != nil {
					log.Fatalln(err)
				}
				os.Exit(0)
			}

			for range time.NewTicker(every).C {
				if err := sync(); err != nil {
					log.Println(err)
				}
			}
		},
	}
	syncCmd.Flags().Bool(`full`, false, `初次备份是否全量扫描更新。`)
	syncCmd.Flags().Duration(`every`, 0, `每隔多久同步一次。如果不设置，默认只同步一次。`)
	parent.AddCommand(syncCmd)
}
