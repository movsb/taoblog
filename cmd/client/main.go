package client

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/goccy/go-yaml"
	client_common "github.com/movsb/taoblog/cmd/client/common"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 如果当前在开发目录下，则默认为 blog.local，否则为 blog。
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

	name := utils.IIF(version.DevMode(), `dev`, `live`)
	hostConfig, ok := hostConfigs[name]
	if !ok {
		panic("cannot find init config for host: " + name)
	}

	return hostConfig
}

func saveHostConfig(name string, home string, token string) {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	path := filepath.Join(usr.HomeDir, "/.taoblog.yml")
	fp, err := os.Open(path)
	if err != nil {
		if !os.IsNotExist(err) {
			panic("cannot read init config: " + path)
		}
	}
	hostConfigs := map[string]HostConfig{}
	ymlDec := yaml.NewDecoder(fp)
	if err := ymlDec.Decode(&hostConfigs); err != nil {
		panic(err)
	}
	fp.Close()

	fp = utils.Must1(os.Create(path))
	defer fp.Close()

	hostConfigs[name] = HostConfig{
		Home:  home,
		Token: token,
	}

	yaml.NewEncoder(fp).Encode(hostConfigs)
}

var config HostConfig
var client *Client

func AddCommands(rootCmd *cobra.Command) {
	loginCmd := &cobra.Command{
		Use:   `login <home url>`,
		Short: `登录到站点`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			homeURL, err := url.Parse(args[0])
			if err != nil {
				log.Fatalln(err)
			}
			// 用于生成登录挑战，不需要 token。
			client := clients.NewFromHome(homeURL.String(), ``)
			beginLogin := utils.Must1(client.Auth.ClientLogin(client.Context(), &proto.ClientLoginRequest{}))
			defer beginLogin.CloseSend()

			save := func(token string) {
				name := utils.IIF(version.DevMode(), `dev`, `live`)
				saveHostConfig(name, homeURL.String(), token)
			}

			for {
				rsp := utils.Must1(beginLogin.Recv())
				if open := rsp.GetOpen(); open != nil {
					fmt.Println(`Open URL:`, open.AuthUrl)
					continue
				}
				if succ := rsp.GetSuccess(); succ != nil {
					save(succ.Token)
					fmt.Println(`Success.`)
					break
				}
			}
		},
	}
	rootCmd.AddCommand(loginCmd)

	preRun := func(cmd *cobra.Command, args []string) {
		config = InitHostConfigs()
		client = NewClient(config)
	}

	createCmd := &cobra.Command{
		Use:              `create`,
		Short:            `快速通过文件创建文章。`,
		PersistentPreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			file := utils.Must1(cmd.Flags().GetString(`file`))
			root := os.DirFS(".")
			file = strings.TrimPrefix(file, "./")

			// 取文件的创建时间
			// 如果形如 年-月-日.md
			//   如果修改时间也在这个日期内，则用修改日期
			//   否则用文件路径日期。
			// 几种格式：
			info := utils.Must1(fs.Stat(root, file))
			fileTime := time.Now()
			// 是否有自己独立的目录名。
			ownDir := false
			var year, month, day, week int
			var nameTime time.Time
			if n, err := fmt.Sscanf(filepath.Base(file), "%d-%d-%d.md", &year, &month, &day); err == nil && n == 3 {
				// 2022/01/2022-01-01.md 表示日期
				nameTime = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
				fileTime = nameTime
			} else if n, err := fmt.Sscanf(file, "%d/%d/%d.md", &year, &month, &day); err == nil && n == 3 {
				// 2022/01/01.md 表示日期
				nameTime = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
				fileTime = nameTime
			} else if n, err := fmt.Sscanf(file, "%d/%d.md", &year, &week); err == nil && n == 2 {
				// 2022/01.md 表示周
				nameTime = time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
				foundWeek := false
				for i := 0; i < 366; i++ {
					_, w := nameTime.ISOWeek()
					if w == week {
						foundWeek = true
						break
					}
					nameTime = nameTime.Add(time.Hour * 24)
				}
				if !foundWeek {
					panic("没找到对应的周。")
				}
				fileTime = nameTime
			} else {
				// 从 git 拿
				cmd := exec.Command(`bash`, `-c`, fmt.Sprintf(`git log --follow --format=%%ad --date default "%s" | tail -1`, file))
				out, err := cmd.Output()
				if err != nil {
					panic(err.Error() + cmd.String())
				}
				// Wed Sep 20 23:06:36 2023 +0800
				t, err := time.Parse(`Mon Jan 2 15:04:05 2006 -0700`, strings.TrimSpace(string(out)))
				if err != nil {
					panic(err.Error() + cmd.String())
				}
				nameTime = t
				fileTime = nameTime
				ownDir = true
			}
			if nameTime.IsZero() {
				fileTime = info.ModTime()
				// } else if info.ModTime().Sub(nameTime) < time.Hour*24*7 {
				// 	fileTime = info.ModTime()
			}
			post := utils.Must1(client.Blog.CreatePost(client.Context(), &proto.Post{
				Date:       int32(fileTime.Unix()),
				Modified:   int32(fileTime.Unix()),
				Title:      filepath.Base(file), // 如果文章里面有标题，会自动覆盖
				SourceType: `markdown`,
				Source:     string(utils.Must1(os.ReadFile(file))),
				Status:     utils.Must1(cmd.Flags().GetString(`status`)),
				Type:       `tweet`,
			}))
			var dir string
			if ownDir {
				dir = filepath.Dir(file)
				if dir == "" || dir == "." {
					dir = fileTime.Format(`2006/01/02`)
				}
				dir += "/" + strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))
			} else {
				dir = fileTime.Format(`2006/01/02`)
				dir += fmt.Sprintf("/%d", post.Id)
			}
			utils.Must(os.MkdirAll(dir, 0755))
			utils.Must(client_common.SavePostConfig(filepath.Join(dir, client_common.ConfigFileName), &client_common.PostConfig{
				ID:       post.Id,
				Title:    post.Title,
				Modified: post.Modified,
				Tags:     post.Tags,
				Metas:    *models.PostMetaFrom(post.Metas),
				Slug:     post.Slug,
				Type:     post.Type,
			}))
			mvCmd := exec.Command(`git`, `mv`, file, filepath.Join(dir, `README.md`))
			utils.Must(mvCmd.Run())
		},
	}
	createCmd.Flags().StringP(`file`, `f`, ``, `文章对应的文件（README.md）`)
	createCmd.Flags().StringP(`status`, `s`, `draft`, `状态（public、draft）`)
	rootCmd.AddCommand(createCmd)

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
			client.SetPostStatus(postID, models.PostStatusPublic, touch)
		},
	}
	postsPublishCmd.Flags().BoolP(`touch`, `t`, false, `Touch create_time and update_time`)
	postsCmd.AddCommand(postsPublishCmd)
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

	postsCreateStylingPageCmd := &cobra.Command{
		Use:   `create-styling-page`,
		Short: `创建样式测试页面`,
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			source := utils.Must1(cmd.Flags().GetString(`source`))
			url, err := url.Parse(source)
			if err == nil && (url.Scheme == `http` || url.Scheme == `https`) {
				rsp, err := http.Get(source)
				if err != nil {
					log.Fatalln(err)
				}
				if rsp.StatusCode != 200 {
					log.Fatalln(`code != 200`)
				}
				doc := utils.Must1(goquery.NewDocumentFromReader(io.LimitReader(rsp.Body, 1<<20)))
				body := doc.Find(`body`)
				if len(body.Nodes) <= 0 {
					log.Fatalln(`no body`)
				}
				source = utils.Must1(body.Html())
				// log.Println(source)
			} else if source != `` {
				source = string(utils.Must1(os.ReadFile(source)))
			}
			utils.Must1(client.Blog.CreateStylingPage(client.Context(), &proto.CreateStylingPageRequest{
				Source: source,
			}))
		},
	}
	postsCreateStylingPageCmd.Flags().StringP(`source`, `s`, ``, `文章源内容路径，支持指定网页。`)
	postsCmd.AddCommand(postsCreateStylingPageCmd)

	postsTransferCmd := &cobra.Command{
		Use:              `transfer <post-id> <user-id>`,
		Short:            `转移文章给用户。`,
		PersistentPreRun: preRun,
		Args:             cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var (
				postID = utils.Must1(strconv.Atoi(args[0]))
				userID = utils.Must1(strconv.Atoi(args[1]))
			)
			utils.Must1(client.Blog.SetPostUserID(
				client.Context(),
				&proto.SetPostUserIDRequest{
					PostId: int64(postID),
					UserId: int32(userID),
				},
			))
		},
	}
	postsCmd.AddCommand(postsTransferCmd)

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
	backupCmd := &cobra.Command{
		Use:              `backup`,
		Short:            `备份文章、评论、文件等`,
		Args:             cobra.NoArgs,
		PersistentPreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			compress := utils.Must1(cmd.Flags().GetBool(`compress`))
			keepLogs := utils.Must1(cmd.Flags().GetBool(`keep-logs`))
			client.Backup(cmd, compress, !keepLogs)
		},
	}
	backupCmd.Flags().Bool(`compress`, true, `是否压缩传输。`)
	backupCmd.Flags().Bool(`keep-logs`, false, `是否保留未处理的日志（通知、邮件等）。`)
	rootCmd.AddCommand(backupCmd)

	configCmd := &cobra.Command{
		Use:              `config`,
		Short:            `get/set config`,
		PersistentPreRun: preRun,
	}
	rootCmd.AddCommand(configCmd)
	configGetCmd := &cobra.Command{
		Use:   `get`,
		Short: `get [/][path.to.config]`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			value, err := client.GetConfig(path)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Println(value)
		},
	}
	configCmd.AddCommand(configGetCmd)
	configSetCmd := &cobra.Command{
		Use:   `set`,
		Short: `set [/]<path.to.config> [value/stdin]`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var (
				path  = args[0]
				value = ""
			)
			if len(args) >= 2 {
				value = args[1]
			} else {
				value = string(utils.Must1(io.ReadAll(os.Stdin)))
			}
			if err := client.SetConfig(path, value); err != nil {
				log.Fatalln(err)
			}
		},
	}
	configCmd.AddCommand(configSetCmd)
	configEditCmd := &cobra.Command{
		Use:   `edit`,
		Short: `edit [/][path.to.config]`,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			value, err := client.GetConfig(path)
			if err != nil {
				if status.Code(err) == codes.NotFound && strings.HasPrefix(path, `/`) {
					value = ""
				} else {
					log.Fatalln(err)
				}
			}
			ext := `.yaml`
			if strings.HasPrefix(path, `/`) {
				ext = filepath.Ext(path)
			}
			for {
				editedValue, edited := edit(value, ext)
				if !edited {
					break
				}
				err := client.SetConfig(path, editedValue)
				if err == nil {
					break
				}
				log.Println(`更新配置时错误：`, err)
				fmt.Print(`按回车重新编辑，Ctrl+C 退出...`)
				if _, err := fmt.Scanln(); err != nil {
					break
				}
				value = editedValue
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

	updateCmd := &cobra.Command{
		Use:    `update`,
		Short:  `计划重启任务标识。`,
		Args:   cobra.NoArgs,
		PreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			client.Update()
		},
	}
	rootCmd.AddCommand(updateCmd)

	infoCmd := &cobra.Command{
		Use:    `info`,
		Short:  ``,
		Args:   cobra.NoArgs,
		PreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			client.Info()
		},
	}
	rootCmd.AddCommand(infoCmd)

	users := createUsersCommands()
	users.PersistentPreRun = preRun
	rootCmd.AddCommand(users)

	proxyCmd := &cobra.Command{
		Use:              `proxy`,
		Short:            `代理网络请求，自动登录。`,
		PersistentPreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			proxy(cmd.Context(), `localhost:22564`, config.Home, config.Token)
		},
	}
	rootCmd.AddCommand(proxyCmd)
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
