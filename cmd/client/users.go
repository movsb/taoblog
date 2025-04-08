package client

import (
	"io"
	"log"
	"os"
	"strconv"

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

	listCmd := &cobra.Command{
		Use:   `list`,
		Short: `列举所有用户`,
		Run: func(cmd *cobra.Command, args []string) {
			hidden := utils.Must1(cmd.Flags().GetBool(`hidden`))
			unnamed := utils.Must1(cmd.Flags().GetBool(`unnamed`))
			u, err := client.Auth.ListUsers(client.Context(),
				&proto.ListUsersRequest{
					WithHidden:  hidden,
					WithUnnamed: unnamed,
				},
			)
			if err != nil {
				log.Fatalln(err)
			}
			yaml.NewEncoder(os.Stdout).Encode(u)
		},
	}
	listCmd.Flags().Bool(`hidden`, false, ``)
	listCmd.Flags().MarkHidden(`hidden`)
	listCmd.Flags().Bool(`unnamed`, false, `包含未使用的`)
	usersCmd.AddCommand(listCmd)

	updateCmd := &cobra.Command{
		Use:   `update <id>`,
		Short: `更新用户`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := utils.Must1(strconv.Atoi(args[0]))

			req := &proto.UpdateUserRequest{
				User: &proto.User{
					Id: int64(id),
				},
			}

			if cmd.Flags().Changed(`avatar`) {
				path := utils.Must1(cmd.Flags().GetString(`avatar`))
				f := utils.Must1(os.Open(path))
				defer f.Close()
				st := utils.Must1(f.Stat())
				if st.Size() > 100<<10 {
					log.Fatalln(`文件太大。`)
				}
				d := utils.CreateDataURL(utils.Must1(io.ReadAll(f)))
				req.User.Avatar = d.String()
				req.UpdateAvatar = true
			}

			utils.Must1(client.Auth.UpdateUser(client.Context(), req))
		},
	}
	updateCmd.Flags().String(`avatar`, ``, `头像文件路径。`)
	usersCmd.AddCommand(updateCmd)

	return usersCmd
}
