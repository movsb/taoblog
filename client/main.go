package main

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"

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

func main() {
	config := initHostConfigs()
	client := NewClient(config)

	rootCmd := &cobra.Command{
		Use: `taoblog`,
	}
	getCmd := &cobra.Command{
		Use: `get`,
		Annotations: map[string]string{
			`K`: `V`,
		},
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client.CRUD(cmd.Use, args[0])
		},
	}
	postCmd := &cobra.Command{
		Use:  `post`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client.CRUD(cmd.Use, args[0])
		},
	}
	deleteCmd := &cobra.Command{
		Use:  `delete`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client.CRUD(cmd.Use, args[0])
		},
	}
	patchCmd := &cobra.Command{
		Use:  `patch`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client.CRUD(cmd.Use, args[0])
		},
	}
	rootCmd.AddCommand(getCmd, postCmd, deleteCmd, patchCmd)
	postsCmd := &cobra.Command{
		Use: `posts`,
	}
	rootCmd.AddCommand(postsCmd)
	postsInitCmd := &cobra.Command{
		Use: `init`,
		Run: func(cmd *cobra.Command, args []string) {
			client.InitPost()
		},
	}
	postCmd.AddCommand(postsInitCmd)
	postsCreateCmd := &cobra.Command{
		Use: `create`,
		Run: func(cmd *cobra.Command, args []string) {
			client.CreatePost()
		},
	}
	postsCmd.AddCommand(postsCreateCmd)
	postsUploadCmd := &cobra.Command{
		Use:  `upload`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client.UploadPostFiles(args)
		},
	}
	postCmd.AddCommand(postsUploadCmd)
	postsUpdateCmd := &cobra.Command{
		Use: `update`,
		Run: func(cmd *cobra.Command, args []string) {
			client.UpdatePost()
		},
	}
	postCmd.AddCommand(postsUpdateCmd)
	postsPublishCmd := &cobra.Command{
		Use:     `publish`,
		Aliases: []string{`pub`},
		Run: func(cmd *cobra.Command, args []string) {
			client.SetPostStatus(`public`)
		},
	}
	postCmd.AddCommand(postsPublishCmd)
	postsDraftCmd := &cobra.Command{
		Use: `draft`,
		Run: func(cmd *cobra.Command, args []string) {
			client.SetPostStatus(`draft`)
		},
	}
	postCmd.AddCommand(postsDraftCmd)
	postsGetCmd := &cobra.Command{
		Use: `get`,
		Run: func(cmd *cobra.Command, args []string) {
			client.GetPost()
		},
	}
	postCmd.AddCommand(postsGetCmd)
	commentsCmd := &cobra.Command{
		Use: `comments`,
	}
	rootCmd.AddCommand(commentsCmd)
	commentsCmd.AddCommand(&cobra.Command{
		Use:  `set-post-id`,
		Args: cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cmtID, err := strconv.ParseInt(os.Args[0], 10, 0)
			if err != nil {
				panic(err)
			}
			postID, err := strconv.ParseInt(os.Args[1], 10, 0)
			if err != nil {
				panic(err)
			}
			client.SetCommentPostID(cmtID, postID)
		},
	})
	commentsCmd.AddCommand(&cobra.Command{
		Use:  `edit`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cmtID, err := strconv.ParseInt(os.Args[3], 10, 0)
			if err != nil {
				panic(err)
			}
			client.UpdateComment(cmtID)
		},
	})
	backupCmd := &cobra.Command{
		Use: `backup`,
		Run: func(cmd *cobra.Command, args []string) {
			client.Backup(os.Stdout)
		},
	}
	rootCmd.AddCommand(backupCmd)
	settingsCmd := &cobra.Command{
		Use:  `settings`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client.Settings(args)
		},
	}
	rootCmd.AddCommand(settingsCmd)

	rootCmd.Execute()
}
