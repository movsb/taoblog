package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"

	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// 如果当前在开发目录下，则默认为 blog.local，否则为 blog。
// 在开发目录时仍然可以用环境变量 LIVE=1 来使用 blog。
func InitHostConfigs() HostConfig {
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
		if _, err := os.Stat(`go.mod`); err == nil {
			if os.Getenv(`LIVE`) != `1` {
				host = `blog.local`
			}
		}
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
		config = InitHostConfigs()
		client = NewClient(config)
	}

	pingCmd := &cobra.Command{
		Use:    `ping`,
		Short:  `Ping server`,
		Args:   cobra.NoArgs,
		PreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			resp, err := client.Blog.Ping(context.Background(), &proto.PingRequest{})
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
			if err := client.InitPost(); err != nil {
				panic(err)
			}
		},
	}
	postsCmd.AddCommand(postsInitCmd)
	postsCreateCmd := &cobra.Command{
		Use:   `create`,
		Short: `Create the post in this directory`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.CreatePost(); err != nil {
				panic(err)
			}
		},
	}
	postsCmd.AddCommand(postsCreateCmd)
	postsUploadCmd := &cobra.Command{
		Use:        `upload <files...>`,
		Short:      `Upload post assets, like images`,
		Args:       cobra.MinimumNArgs(1),
		Deprecated: `将会自动上传文章附件，此命令不再需要手动执行。`,
		Run: func(cmd *cobra.Command, args []string) {
			p := client.readPostConfig()
			if p.ID <= 0 {
				panic("未发表的文章")
			}
			client.UploadPostFiles(p.ID, args)
		},
	}
	postsCmd.AddCommand(postsUploadCmd)
	postsUpdateCmd := &cobra.Command{
		Use:   `update`,
		Short: `Update post in this directory`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := client.UpdatePost(); err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err.Error())
				os.Exit(1)
			}
		},
	}
	postsCmd.AddCommand(postsUpdateCmd)
	postApplyCmd := &cobra.Command{
		Use:   `apply`,
		Short: `Init, Create and Update a post`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			switch err := client.InitPost(); err {
			case nil:
				return
			case errPostInited:
				break
			default:
				panic(err)
			}

			switch err := client.CreatePost(); err {
			case nil:
				return
			case errPostCreated:
				break
			default:
				panic(err)
			}

			if err := client.UpdatePost(); err != nil {
				fmt.Fprintf(os.Stderr, "update failed: %v\n", err.Error())
				os.Exit(1)
			}
		},
	}
	postsCmd.AddCommand(postApplyCmd)
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
			touch, err := cmd.Flags().GetBool(`touch`)
			if err != nil {
				panic(err)
			}
			client.SetPostStatus(postID, true, touch)
		},
	}
	postsPublishCmd.Flags().BoolP(`touch`, `t`, false, `Touch create_time and update_time`)
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
			client.SetPostStatus(postID, false, false)
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
	postsDeleteCmd := &cobra.Command{
		Use:   `delete <post-id>`,
		Short: `Delete a post`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			postID, err := strconv.ParseInt(args[0], 10, 0)
			if err != nil {
				panic(err)
			}
			if err := client.DeletePost(postID); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	postsCmd.AddCommand(postsDeleteCmd)

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
		Short:            `Backup ...`,
		Args:             cobra.NoArgs,
		PersistentPreRun: preRun,
	}
	rootCmd.AddCommand(backupCmd)
	backupPostsCmd := &cobra.Command{
		Use:   `posts`,
		Short: `backup posts and comments`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client.BackupPosts(cmd)
		},
	}
	backupPostsCmd.Flags().Bool(`stdout`, false, `Output to stdout`)
	backupPostsCmd.Flags().Bool(`no-link`, false, `Don't link to taoblog.db`)
	backupCmd.AddCommand(backupPostsCmd)
	backupFilesCmd := &cobra.Command{
		Use:   `files`,
		Short: `backup files`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client.BackupFiles(cmd)
		},
	}
	backupCmd.AddCommand(backupFilesCmd)

	configCmd := &cobra.Command{
		Use:              `config`,
		Short:            `get/set config`,
		PersistentPreRun: preRun,
	}
	rootCmd.AddCommand(configCmd)
	configGetCmd := &cobra.Command{
		Use:   `get`,
		Short: `get [path.to.config]`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			value := client.GetConfig(path)
			fmt.Println(value)
		},
	}
	configCmd.AddCommand(configGetCmd)
	configSetCmd := &cobra.Command{
		Use:   `set`,
		Short: `set <path.to.config> value`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			path, value := args[0], args[1]
			client.SetConfig(path, value)
		},
	}
	configCmd.AddCommand(configSetCmd)
	configEditCmd := &cobra.Command{
		Use:   `edit`,
		Short: `edit [path.to.config]`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			value := client.GetConfig(path)
			if editedValue, edited := edit(value, `.yaml`); edited {
				client.SetConfig(path, editedValue)
			}
		},
	}
	configCmd.AddCommand(configEditCmd)

	restartCmd := &cobra.Command{
		Use:    `restart`,
		Args:   cobra.NoArgs,
		PreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			client.Restart()
		},
	}
	rootCmd.AddCommand(restartCmd)

	lfs := createLfsCommands()
	lfs.PersistentPreRun = preRun
	rootCmd.AddCommand(lfs)
}

func edit(value string, fileSuffix string) (string, bool) {
	editor, ok := os.LookupEnv(`EDITOR`)
	if !ok {
		editor = `vim`
	}

	tmpFile, err := ioutil.TempFile(``, `taoblog-edit-*`+fileSuffix)
	if err != nil {
		panic(err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(value); err != nil {
		panic(err)
	}

	oldInfo, err := tmpFile.Stat()
	if err != nil {
		panic(err)
	}

	tmpFile.Close()

	// fmt.Printf("Editing: %d, post: %d\n", cmt.Id, cmt.PostId)

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}

	newInfo, err := os.Stat(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	if newInfo.ModTime() == oldInfo.ModTime() {
		fmt.Println(`file not modified`)
		return value, false
	}

	bys, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		panic(err)
	}

	return string(bys), true
}
