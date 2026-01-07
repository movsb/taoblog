package client

import (
	"io"
	"log"
	"os"
	"strconv"

	"github.com/goccy/go-yaml"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/spf13/cobra"
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
			hidden := utils.Must1(cmd.Flags().GetBool(`hidden`))
			u, err := client.Users.CreateUser(client.Context(),
				&proto.User{
					Nickname: nickname,
					Hidden:   hidden,
				},
			)
			if err != nil {
				log.Fatalln(err)
			}
			yaml.NewEncoder(os.Stdout).Encode(u)
		},
	}
	createCmd.Flags().StringP(`nickname`, `n`, ``, `昵称（不能为空）`)
	createCmd.Flags().Bool(`hidden`, false, `归档用户，不对外显示。`)
	createCmd.Flags().MarkHidden(`hidden`)
	usersCmd.AddCommand(createCmd)

	listCmd := &cobra.Command{
		Use:   `list`,
		Short: `列举所有用户`,
		Run: func(cmd *cobra.Command, args []string) {
			hidden := utils.Must1(cmd.Flags().GetBool(`hidden`))
			unnamed := utils.Must1(cmd.Flags().GetBool(`unnamed`))
			u, err := client.Users.ListUsers(client.Context(),
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

			utils.Must1(client.Users.UpdateUser(client.Context(), req))
		},
	}
	updateCmd.Flags().String(`avatar`, ``, `头像文件路径。`)
	usersCmd.AddCommand(updateCmd)

	setCmd := &cobra.Command{
		Use:   `set`,
		Short: `更新用户参数。`,
		Run: func(cmd *cobra.Command, args []string) {
			v := proto.SetUserSettingsRequest{
				Settings: &proto.Settings{},
			}
			if changed := cmd.Flags().Changed(`review-posts-in-calendar`); changed {
				v.Settings.ReviewPostsInCalendar = utils.Must1(cmd.Flags().GetBool(`review-posts-in-calendar`))
				v.UpdateReviewPostsInCalendar = true
			}
			utils.Must1(client.Blog.SetUserSettings(client.Context(), &v))
		},
	}
	setCmd.Flags().Bool(`review-posts-in-calendar`, false, `是否在日历中回顾进入文章。`)
	usersCmd.AddCommand(setCmd)

	getCmd := &cobra.Command{
		Use:   `get`,
		Short: `获取用户参数。`,
		Run: func(cmd *cobra.Command, args []string) {
			s := utils.Must1(client.Blog.GetUserSettings(client.Context(), &proto.GetUserSettingsRequest{}))
			yaml.NewEncoder(os.Stdout).Encode(s)
		},
	}
	usersCmd.AddCommand(getCmd)
	return usersCmd
}
