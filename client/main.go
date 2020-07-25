package client

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/movsb/taoblog/protocols"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func initHostConfigs() HostConfig {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(usr.HomeDir, "/.taoblog.yml")
	fp, err := os.Open(path)
	if err != nil {
		panic("cannot read init config: " + path)
	}
	defer fp.Close()

	hostConfigs := map[string]HostConfig{}
	ymlDec := yaml.NewDecoder(fp)
	if err := ymlDec.Decode(&hostConfigs); err != nil {
		panic(err)
	}

	// select which host to use
	host := os.Getenv("HOST")
	if host == "" {
		host = "blog"
	}
	hostConfig, ok := hostConfigs[host]
	if !ok {
		panic("cannot find init config for host: " + host)
	}
	return hostConfig
}

var config HostConfig
var client *Client

// AddCommands ...
func AddCommands(rootCmd *cobra.Command) {
	preRun := func(cmd *cobra.Command, args []string) {
		config = initHostConfigs()
		client = NewClient(config)
	}

	pingCmd := &cobra.Command{
		Use:    `ping`,
		Short:  `Ping server`,
		Args:   cobra.NoArgs,
		PreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.grpcClient.Ping(context.Background(), &protocols.PingRequest{})
			if err != nil {
				panic(err)
			}
			fmt.Println(resp.Pong)
		},
	}
	rootCmd.AddCommand(pingCmd)
	postsCmd := &cobra.Command{
		Use:              `posts`,
		Short:            `Commands for managing posts`,
		PersistentPreRun: preRun,
	}
	rootCmd.AddCommand(postsCmd)
	postsInitCmd := &cobra.Command{
		Use:   `init`,
		Short: `Initialize an empty post structure in this directory`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client.InitPost()
		},
	}
	postsCmd.AddCommand(postsInitCmd)
	postsCreateCmd := &cobra.Command{
		Use:   `create`,
		Short: `Create the post in this directory`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client.CreatePost()
		},
	}
	postsCmd.AddCommand(postsCreateCmd)
	postsUploadCmd := &cobra.Command{
		Use:   `upload`,
		Short: `Upload post assets, like images`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client.UploadPostFiles(args)
		},
	}
	postsCmd.AddCommand(postsUploadCmd)
	postsUpdateCmd := &cobra.Command{
		Use:   `update`,
		Short: `Update post in this directory`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client.UpdatePost()
		},
	}
	postsCmd.AddCommand(postsUpdateCmd)
	postsPublishCmd := &cobra.Command{
		Use:     `publish [post-id]`,
		Short:   `Publish this post`,
		Aliases: []string{`pub`},
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var postID int64
			var err error
			if len(args) == 1 {
				postID, err = strconv.ParseInt(args[0], 10, 0)
				if err != nil {
					panic(err)
				}
			}
			client.SetPostStatus(postID, true)
		},
	}
	postsCmd.AddCommand(postsPublishCmd)
	postsDraftCmd := &cobra.Command{
		Use:   `draft [post-id]`,
		Short: `Unpublish this post (make it a draft)`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var postID int64
			var err error
			if len(args) == 1 {
				postID, err = strconv.ParseInt(args[0], 10, 0)
				if err != nil {
					panic(err)
				}
			}
			client.SetPostStatus(postID, false)
		},
	}
	postsCmd.AddCommand(postsDraftCmd)
	postsGetCmd := &cobra.Command{
		Use:   `get`,
		Short: `(Don't use)`,
		Run: func(cmd *cobra.Command, args []string) {
			client.GetPost()
		},
	}
	postsCmd.AddCommand(postsGetCmd)
	commentsCmd := &cobra.Command{
		Use:              `comments`,
		Short:            `Commands for managing comments`,
		PersistentPreRun: preRun,
	}
	rootCmd.AddCommand(commentsCmd)
	commentsCmd.AddCommand(&cobra.Command{
		Use:   `set-post-id <comment-id> <post-id>`,
		Short: `Transfer comment to another post`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cmtID, err := strconv.ParseInt(args[0], 10, 0)
			if err != nil {
				panic(err)
			}
			postID, err := strconv.ParseInt(args[1], 10, 0)
			if err != nil {
				panic(err)
			}
			client.SetCommentPostID(cmtID, postID)
		},
	})
	commentsCmd.AddCommand(&cobra.Command{
		Use:   `edit`,
		Short: `Edit some comment`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmtID, err := strconv.ParseInt(os.Args[3], 10, 0)
			if err != nil {
				panic(err)
			}
			client.UpdateComment(cmtID)
		},
	})
	backupCmd := &cobra.Command{
		Use:              `backup`,
		Short:            `Dump database`,
		Args:             cobra.NoArgs,
		PersistentPreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			client.Backup(cmd)
		},
	}
	backupCmd.Flags().Bool(`stdout`, false, `Output to stdout`)
	rootCmd.AddCommand(backupCmd)
}
