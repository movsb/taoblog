package sync

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
)

// 多少时间范围内的更新不应该被检测到。或者说：
// 停止更新多少时间后才算作修改过、才能保存。
// 这样可以防止保存正在频繁编辑的文章。
const skewedDurationForUpdating = time.Minute * 15

type GitSync struct {
	proto *clients.ProtoClient
	root  string

	// 上一次获取更新的时间。
	// 下一次获取时从此时间继续增量获取。
	// NOTE：第一次的可以设置得很远，以纠正中途备份中断的缺失内容，如果有的话。
	// NOTE：本时间不等于计划任务执行的时间点，而是上一次的 notAfter 时间点。
	// NOTE：这样可以保证无论计划任务执行的频次如何，总是能保证 [notBefore, notAfter) 时间有效。
	lastCheckedAt time.Time

	credential Credential
	auth       transport.AuthMethod
}

type Credential struct {
	Author   string
	Email    string
	Username string
	Password string
}

// full: 初次备份是否需要全量扫描备份。如果不设置，则默认为最近 7 天。
func New(config client.HostConfig, credential Credential, root string, full bool) *GitSync {
	client := clients.NewProtoClient(config.Home, config.Token)

	lastCheckedAt := time.Unix(0, 0)
	if !full {
		lastCheckedAt = time.Now().Add(-7 * time.Hour * 24)
	}

	return &GitSync{
		proto: client,
		root:  root,

		credential: credential,
		auth: &http.BasicAuth{
			Username: credential.Username,
			Password: credential.Password,
		},

		lastCheckedAt: lastCheckedAt,
	}
}

func (g *GitSync) Sync() error {
	repo, err := git.PlainOpen(g.root)
	if err != nil {
		return err
	}
	wt, err := repo.Worktree()
	if err != nil {
		return err
	}
	if err := wt.Pull(&git.PullOptions{
		RemoteName:    `origin`,
		ReferenceName: `refs/heads/master`,
		SingleBranch:  true,
		Progress:      os.Stdout,
		Auth:          g.auth,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		log.Println(`Pull 失败：`, err)
		return err
	}

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

	for _, post := range posts {
		// log.Println(`处理：`, post.Id, post.Title)
		if err := g.syncSingle(wt, post); err != nil {
			log.Println(err)
			continue
		}
	}
	log.Println(`共有`, len(posts), `篇文章被处理。`)

	if err := repo.Push(&git.PushOptions{
		RemoteName: `origin`,
		RefSpecs: []config.RefSpec{
			`refs/heads/master:refs/heads/master`,
		},
		Auth: g.auth,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		log.Println(`Push 失败：`, err)
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

func (g *GitSync) syncSingle(wt *git.Worktree, p *proto.Post) error {
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

	// 正在编辑且并没提交的文件可能会比远程更新，此时不能覆盖本地的文件。
	// TODO 用文件系统而不是 os.
	if stat, err := os.Stat(fullPath); err == nil {
		if stat.ModTime().After(time.Unix(int64(p.Modified), 0)) {
			return fmt.Errorf(`本地的文件更新，没有覆盖：%s`, path)
		}
	}

	if err := ioutil.WriteFile(fullPath, []byte(p.Source), 0644); err != nil {
		return err
	}

	log.Println(`正在写入更新：`, fullPath)

	if _, err := wt.Add("."); err != nil {
		log.Println(`git add 失败：`, err)
		return err
	}
	if _, err := wt.Commit(`Updated by Sync command.`, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.credential.Author,
			Email: g.credential.Email,
			When:  time.Unix(int64(p.Modified), 0),
		},
	}); err != nil {
		log.Println(`提交失败：`, err)
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
// TODO：创建索引。
func findPostByID(fsys fs.FS, id int32) (outPath string, outConfig *client.PostConfig, outErr error) {
	err := fs.WalkDir(fsys, `.`, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if d.Name() == client.ConfigFileName {
			fp := utils.Must1(fsys.Open(path))
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
