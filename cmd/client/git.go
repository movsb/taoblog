package client

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
	git_config "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
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

	config *proto.GetSyncConfigResponse
	auth   transport.AuthMethod
}

// full: 初次备份是否需要全量扫描备份。如果不设置，则默认为最近 7 天。
func New(client *clients.ProtoClient, root string, full bool) *GitSync {
	lastCheckedAt := time.Unix(0, 0)
	if !full {
		lastCheckedAt = time.Now().Add(-7 * time.Hour * 24)
	}

	return &GitSync{
		proto: client,
		root:  root,

		lastCheckedAt: lastCheckedAt,
	}
}

// 内部会自动大量重试因为网络问题导致的错误。
func (g *GitSync) Sync() error {
	const MaxRetry = 20

	repo, tree, err := g.prepare()
	if err != nil {
		return err
	}

	for n := 1; ; n++ {
		if err := g.pull(tree); err != nil {
			if n >= MaxRetry {
				return err
			}
			log.Println(err)
			time.Sleep(time.Second * 3)
			continue
		}
		break
	}

	notAfter, err := g.process(tree)
	if err != nil {
		return err
	}

	for n := 1; ; n++ {
		if err := g.push(repo); err != nil {
			if n >= MaxRetry {
				return err
			}
			log.Println(err)
			time.Sleep(time.Second * 3)
			continue
		}
		break
	}

	// 仅在全部成功后更新上次检测的时间。
	g.lastCheckedAt = notAfter

	return nil
}

func (g *GitSync) prepare() (*git.Repository, *git.Worktree, error) {
	config, err := g.proto.Management.GetSyncConfig(g.proto.Context(), &proto.GetSyncConfigRequest{})
	if err != nil {
		return nil, nil, err
	}

	g.config = config
	g.auth = &http.BasicAuth{
		Username: config.Username,
		Password: config.Password,
	}

	repo, err := git.PlainOpen(g.root)
	if err != nil {
		return nil, nil, fmt.Errorf(`failed to git open: %w`, err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		return nil, nil, err
	}
	return repo, wt, nil
}

func (g *GitSync) pull(tree *git.Worktree) error {
	if err := tree.Pull(&git.PullOptions{
		RemoteName:    `origin`,
		ReferenceName: `refs/heads/master`,
		SingleBranch:  true,
		Progress:      os.Stdout,
		Auth:          g.auth,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf(`failed to git pull: %w`, err)
	}
	return nil
}

func (g *GitSync) process(tree *git.Worktree) (time.Time, error) {
	notBefore := g.lastCheckedAt
	notAfter := time.Now().Add(-skewedDurationForUpdating)

	posts, err := g.getUpdatedPosts(notBefore, notAfter)
	if err != nil {
		return time.Time{}, fmt.Errorf(`获取列表失败：%w`, err)
	}
	if len(posts) == 0 {
		return notAfter, nil
	}

	for _, post := range posts {
		// log.Println(`处理：`, post.Id, post.Title)
		if err := g.syncSingle(tree, post); err != nil {
			log.Println(err)
			continue
		}
	}
	log.Println(`共有`, len(posts), `篇文章被处理。`)
	return notAfter, nil
}

func (g *GitSync) push(repo *git.Repository) error {
	if err := repo.Push(&git.PushOptions{
		RemoteName: `origin`,
		RefSpecs: []git_config.RefSpec{
			`refs/heads/master:refs/heads/master`,
		},
		Auth: g.auth,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf(`failed to git push: %w`, err)
	}
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
		path = filepath.Join(dir, ConfigFileName)
		config = &PostConfig{}
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
	if err := SavePostConfig(filepath.Join(g.root, path), config); err != nil {
		return err
	}
	// TODO 没用 fsys。
	fullPath := filepath.Join(g.root, filepath.Dir(path), IndexFileName)

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
			Name:  g.config.Author,
			Email: g.config.Email,
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
func findPostByID(fsys fs.FS, id int32) (outPath string, outConfig *PostConfig, outErr error) {
	err := fs.WalkDir(fsys, `.`, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if d.Name() == ConfigFileName {
			fp := utils.Must1(fsys.Open(path))
			defer fp.Close()
			c, err := ReadPostConfigReader(fp)
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
