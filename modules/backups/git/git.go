package backups_git

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	git_config "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	client_common "github.com/movsb/taoblog/cmd/client/common"
	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/clients"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	runtime_config "github.com/movsb/taoblog/service/modules/runtime"
)

// 多少时间范围内的更新不应该被检测到。或者说：
// 停止更新多少时间后才算作修改过、才能保存。
// 这样可以防止保存正在频繁编辑的文章。
const skewedDurationForUpdating = time.Minute * 15

type GitSync struct {
	ctx    context.Context
	rc     *RuntimeConfig
	client *clients.ProtoClient

	// 上一次获取更新的时间。
	// 下一次获取时从此时间继续增量获取。
	// NOTE：第一次的可以设置得很远，以纠正中途备份中断的缺失内容，如果有的话。
	// NOTE：本时间不等于计划任务执行的时间点，而是上一次的 notAfter 时间点。
	// NOTE：这样可以保证无论计划任务执行的频次如何，总是能保证 [notBefore, notAfter) 时间有效。
	lastCheckedAt time.Time

	pathCache map[int]string

	config *proto.GetSyncConfigResponse
	auth   transport.AuthMethod

	// 临时保存仓库的目录。
	tmpDir string
}

type RuntimeConfig struct {
	SyncNow bool `yaml:"sync_now"`

	syncNow chan bool

	config.Saver
}

func (c *RuntimeConfig) AfterSet(paths config.Segments, obj any) {
	switch paths.At(0).Key {
	case `sync_now`:
		c.syncNow <- obj.(bool)
	}
}

// full: 初次备份是否需要全量扫描备份。如果不设置，则默认为最近 7 天。
// ctx 用于进程控制，client 应包含凭证。
func New(ctx context.Context, client *clients.ProtoClient, full bool) *GitSync {
	lastCheckedAt := time.Unix(0, 0)
	if !full {
		lastCheckedAt = time.Now().Add(-7 * time.Hour * 24)
	}

	rc := &RuntimeConfig{
		syncNow: make(chan bool),
	}
	if r := runtime_config.FromContext(ctx); r != nil {
		r.Register(`git`, rc)
	}

	return &GitSync{
		ctx:    ctx,
		rc:     rc,
		client: client,

		lastCheckedAt: lastCheckedAt,
		pathCache:     map[int]string{},
	}
}

func (g *GitSync) Do() <-chan bool {
	return g.rc.syncNow
}

// 内部会自动大量重试因为网络问题导致的错误。
func (g *GitSync) Sync() error {
	const MaxRetry = 20

	repo, tree, err := g.prepare()
	if err != nil {
		return err
	}

	for n := 1; ; n++ {
		if err := g.pull(repo, tree); err != nil {
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

func (g *GitSync) prepare() (_ *git.Repository, _ *git.Worktree, outErr error) {
	defer utils.CatchAsError(&outErr)

	config := utils.Must1(g.client.Management.GetSyncConfig(
		g.client.ContextFrom(g.ctx),
		&proto.GetSyncConfigRequest{}),
	)

	g.config = config
	g.auth = &http.BasicAuth{
		Username: config.Username,
		Password: config.Password,
	}

	var repo *git.Repository

	if g.tmpDir == `` {
		log.Println(`正在克隆仓库：`, config.Url)
		g.tmpDir, repo = utils.Must2(clone(g.ctx, config.Url, g.auth))
	} else {
		log.Println(`使用已有仓库：`, g.tmpDir)
		repo = utils.Must1(git.PlainOpen(g.tmpDir))
	}

	return repo, utils.Must1(repo.Worktree()), nil
}

func clone(ctx context.Context, url string, auth transport.AuthMethod) (_ string, _ *git.Repository, outErr error) {
	defer utils.CatchAsError(&outErr)

	dir := utils.Must1(os.MkdirTemp(``, fmt.Sprintf(`%s-git-sync-*`, version.NameLowercase)))

	repo := utils.Must1(git.PlainCloneContext(ctx, dir, false, &git.CloneOptions{
		URL:           url,
		Auth:          auth,
		RemoteName:    `origin`,
		ReferenceName: `main`,
		SingleBranch:  true,
		Depth:         1,
		Progress:      os.Stdout,
	}))

	return dir, repo, nil
}

// fetch & reset --hard
func (g *GitSync) pull(repo *git.Repository, tree *git.Worktree) (outErr error) {
	defer utils.CatchAsError(&outErr)

	remoteBranch := `refs/remotes/origin/main`

	if err := repo.FetchContext(
		context.Background(),
		&git.FetchOptions{
			RemoteName: `origin`,
			RefSpecs: []git_config.RefSpec{
				git_config.RefSpec(`+refs/heads/main:` + remoteBranch),
			},
			Auth:     g.auth,
			Progress: os.Stdout,
			Force:    true,
		},
	); err != nil && err != git.NoErrAlreadyUpToDate {
		panic(fmt.Errorf(`failed to git fetch: %w`, err))
	}

	remoteHead := utils.Must1(repo.Reference(plumbing.ReferenceName(remoteBranch), true))

	utils.Must(tree.Reset(&git.ResetOptions{
		Commit: remoteHead.Hash(),
		Mode:   git.HardReset,
	}))

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
			`refs/heads/main:refs/heads/main`,
		},
		Auth: g.auth,
	}); err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf(`failed to git push: %w`, err)
	}
	return nil
}

// 根据日期创建文章对应的目录。
func (g *GitSync) createPostDir(t int32, id int64) (string, error) {
	createdAt := time.Unix(int64(t), 0).Local()
	dir := createdAt.Format(`2006/01`)
	dir = filepath.Join(dir, fmt.Sprint(id))
	fullDir := filepath.Join(g.tmpDir, dir)
	if err := os.MkdirAll(fullDir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

func (g *GitSync) syncSingle(wt *git.Worktree, p *proto.Post) (outErr error) {
	defer utils.CatchAsError(&outErr)

	path, config, err := findPostByID(os.DirFS(g.tmpDir), int32(p.Id), g.pathCache)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		dir, err := g.createPostDir(p.Date, p.Id)
		if err != nil {
			return err
		}
		path = filepath.Join(dir, client_common.ConfigFileName)
		config = &client_common.PostConfig{}
	}

	config.ID = p.Id
	config.Metas = *models.PostMetaFrom(p.Metas)
	config.Modified = p.Modified
	config.Slug = p.Slug
	config.Tags = p.Tags
	config.Title = p.Title
	config.Type = p.Type

	utils.Must(client_common.SavePostConfig(filepath.Join(g.tmpDir, path), config))

	// TODO 没用 fsys。
	fullPath := filepath.Join(g.tmpDir, filepath.Dir(path), client_common.IndexFileName)

	utils.Must(os.WriteFile(fullPath, []byte(p.Source), 0644))

	log.Println(`正在写入更新：`, fullPath)

	utils.Must1(wt.Add(`.`))

	if _, err := wt.Commit(`Updated by Sync task.`, &git.CommitOptions{
		Author: &object.Signature{
			Name:  g.config.Author,
			Email: g.config.Email,
			When:  time.Unix(int64(p.Modified), 0),
		},
	}); err != nil {
		if !errors.Is(err, git.ErrEmptyCommit) {
			log.Println(`提交失败：`, err)
			return err
		}
	}

	return nil
}

// 从服务器上获取指定周期内更新过的文章。
// TODO 可选触发立即调用（比如强制归档保留版本的需求），而不是被动周期触发。
// NOTE：时间范围是：
func (g *GitSync) getUpdatedPosts(notBefore, notAfter time.Time) ([]*proto.Post, error) {
	rsp, err := g.client.Blog.ListPosts(
		g.client.ContextFrom(g.ctx),
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
func findPostByID(fsys fs.FS, id int32, cache map[int]string) (string, *client_common.PostConfig, error) {
	if p, ok := cache[int(id)]; ok {
		fp := utils.Must1(fsys.Open(p))
		defer fp.Close()
		c, err := client_common.ReadPostConfigReader(fp)
		if err != nil {
			return ``, nil, err
		}
		if c.ID != int64(id) {
			return ``, nil, fmt.Errorf(`缓存不正确`)
		}
		log.Println(`找到缓存路径：`, id, p)
		return p, c, nil
	}

	var outPath string
	var outConfig *client_common.PostConfig

	err := fs.WalkDir(fsys, `.`, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.Type().IsRegular() {
			return nil
		}
		if d.Name() != client_common.ConfigFileName {
			return nil
		}
		fp := utils.Must1(fsys.Open(path))
		defer fp.Close()
		c, err := client_common.ReadPostConfigReader(fp)
		if err != nil {
			return err
		}
		if c.ID != int64(id) {
			return nil
		}
		outPath = path
		outConfig = c
		return fs.SkipAll
	})
	if err != nil {
		return ``, nil, err
	}

	if outConfig == nil {
		return ``, nil, fmt.Errorf("文章未找到：%d, %w", id, os.ErrNotExist)
	}

	cache[int(id)] = outPath
	return outPath, outConfig, nil
}
