package client

import (
	"log"
	"os"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func createUsersCommands() *cobra.Command {
	usersCmd := &cobra.Command{
		Use:   `users`,
		Short: `用户管理命令。`,
	}

	createCmd := &cobra.Command{
		Use:   `create`,
		Short: `创建用户命令`,
		Run: func(cmd *cobra.Command, args []string) {
			nickname := utils.Must1(cmd.Flags().GetString(`nickname`))
			u, err := client.Auth.CreateUser(client.Context(),
				&proto.User{
					Nickname: nickname,
				},
			)
			if err != nil {
				log.Fatalln(err)
			}
			yaml.NewEncoder(os.Stdout).Encode(u)
		},
	}
	createCmd.Flags().StringP(`nickname`, `n`, ``, `昵称（不能为空）`)
	usersCmd.AddCommand(createCmd)

	return usersCmd
}
