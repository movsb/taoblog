package sync

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/modules/utils"
	proto "github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/service/models"
)

// 多少时间范围内的更新不应该被检测到。或者说：
// 停止更新多少时间后才算作修改过、才能保存。
// 这样可以防止保存正在频繁编辑的文章。
const skewedDurationForUpdating = time.Hour

type GitSync struct {
	proto *clients.ProtoClient
	root  string

	// 上一次获取更新的时间。
	// 下一次获取时从此时间继续增量获取。
	// NOTE：第一次的可以设置得很远，以纠正中途备份中断的缺失内容，如果有的话。
	lastCheckedAt time.Time
}

func New(config client.HostConfig, root string) *GitSync {
	client := clients.NewProtoClient(
		clients.NewConn(config.API, config.GRPC),
		config.Token,
	)
	return &GitSync{
		proto:         client,
		root:          root,
		lastCheckedAt: time.Now().Add(-7 * time.Hour * 24),
	}
}

func (g *GitSync) Sync() error {
	notBefore := g.lastCheckedAt
	notAfter := time.Now().Add(-skewedDurationForUpdating)

	posts, err := g.getUpdatedPosts(notBefore, notAfter)
	if err != nil {
		return fmt.Errorf(`获取列表失败：%w`, err)
	}
	if len(posts) == 0 {
		log.Println(`无文章更新。`)
		return nil
	}

	if err := spawn(`git`, []string{`pull`, `-r`, `--autostash`}, g.root, ``); err != nil {
		return err
	}

	for _, post := range posts {
		// log.Println(`处理：`, post.Id, post.Title)
		if err := g.syncSingle(post); err != nil {
			log.Println(err)
			continue
		}
	}

	if err := spawn(`git`, []string{`push`}, g.root, ``); err != nil {
		return err
	}

	// 仅在全部成功后更新上次检测的时间。
	g.lastCheckedAt = notAfter

	return nil
}

func (g *GitSync) createPostDir(t int32, id int64) (string, error) {
	createdAt := time.Unix(int64(t), 0).Local()
	dir := createdAt.Format(`2006/01/02`)
	dir = filepath.Join(dir, fmt.Sprint(id))
	fullDir := filepath.Join(g.root, dir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func (g *GitSync) syncSingle(p *proto.Post) error {
	path, config, err := findPostByID(os.DirFS(g.root), int32(p.Id))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		dir, err := g.createPostDir(p.Date, p.Id)
		if err != nil {
			return err
		}
		path = filepath.Join(dir, client.ConfigFileName)
		config = &client.PostConfig{}
	}
	if p.Modified == config.Modified {
		return nil
	}
	if p.Modified < config.Modified {
		return fmt.Errorf(`本地比远程文件更新？：%v`, path)
	}
	config.ID = p.Id
	config.Metas = *models.PostMetaFrom(p.Metas)
	config.Modified = p.Modified
	config.Slug = p.Slug
	config.Tags = p.Tags
	config.Title = p.Title
	config.Type = p.Type
	if err := client.SavePostConfig(filepath.Join(g.root, path), config); err != nil {
		return err
	}
	// TODO 没用 fsys。
	fullPath := filepath.Join(g.root, filepath.Dir(path), client.IndexFileName)
	if err := ioutil.WriteFile(fullPath, []byte(p.Source), 0644); err != nil {
		return err
	}
	log.Println(`正在写入更新：`, fullPath)
	date := time.Unix(int64(p.Modified), 0).Local().Format(`2006-01-02 15:04:05`)
	script := fmt.Sprintf(`
set -eu
git add .
git commit -m 'Updated by Sync Tool.' --date='%s'
	`, date)
	return spawn(`bash`, nil, filepath.Dir(fullPath), script)
}

func spawn(name string, args []string, dir string, input string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(input)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// 从服务器上获取指定周期内更新过的文章。
// TODO 可选触发立即调用（比如强制归档保留版本的需求），而不是被动周期触发。
// NOTE：时间范围是：
func (g *GitSync) getUpdatedPosts(notBefore, notAfter time.Time) ([]*proto.Post, error) {
	rsp, err := g.proto.Blog.ListPosts(
		g.proto.Context(),
		&proto.ListPostsRequest{
			ModifiedNotBefore: int32(notBefore.Unix()),
			ModifiedNotAfter:  int32(notAfter.Unix()),
		},
	)
	if err != nil {
		return nil, err
	}
	return rsp.Posts, nil
}

// 根据 ID 找到在本地仓库的路径。
func findPostByID(fsys fs.FS, id int32) (outPath string, outConfig *client.PostConfig, outErr error) {
	err := fs.WalkDir(fsys, `.`, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if d.Name() == client.ConfigFileName {
			fp := utils.Must(fsys.Open(path))
			defer fp.Close()
			c, err := client.ReadPostConfigReader(fp)
			if err != nil {
				log.Fatalln(path, err)
			}
			if c.ID == int64(id) {
				outPath = path
				outConfig = c
				outErr = nil
				return fs.SkipAll
			}
		} else if d.Name() == `metas` {
			fp := utils.Must(fsys.Open(path))
			defer fp.Close()
			all, err := io.ReadAll(fp)
			if err != nil {
				return err
			}
			var id2 int
			if n, err := fmt.Sscanf(string(all), "id:%d", &id2); err != nil || n != 1 {
				return fmt.Errorf("错误的元文件：%s", path)
			}
			if id2 == int(id) {
				// return fmt.Errorf(`旧元数据文件，不知道如何处理: %v`, path)
				// TODO 没有用 fs
				outPath = path
				outConfig = &client.PostConfig{ID: int64(id2)}
				if err := client.SavePostConfig(path, outConfig); err != nil {
					outErr = err
					return fs.SkipAll
				}
				outErr = nil
				return fs.SkipAll
			}
		}
		return nil
	})
	if err != nil {
		outErr = err
		return
	}
	outErr = nil
	if outPath == "" {
		outErr = fmt.Errorf("文章未找到：%d, %w", id, os.ErrNotExist)
	}
	return
}
