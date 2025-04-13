package service

import (
	"context"
	"database/sql"
	"embed"
	"encoding/base64"
	"io/fs"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/gateway/addons"
	"github.com/movsb/taoblog/gateway/handlers/favicon"
	"github.com/movsb/taoblog/gateway/handlers/roots"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/cache"
	commentgeo "github.com/movsb/taoblog/service/modules/comment_geo"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/renderers/exif"
	"github.com/movsb/taoblog/service/modules/renderers/friends"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
	"github.com/movsb/taoblog/service/modules/renderers/rss"
	runtime_config "github.com/movsb/taoblog/service/modules/runtime"
	"github.com/movsb/taoblog/service/modules/search"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"github.com/movsb/taorm"
	"github.com/phuslu/lru"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ToBeImplementedByRpc interface {
	ListAllPostsIds(ctx context.Context) ([]int32, error)
	GetDefaultIntegerOption(name string, def int64) int64
	GetLink(ID int64) string
	GetPlainLink(ID int64) string
	Config() *config.Config
	IncrementViewCount(m map[int]int)
	GetPostsByTags(ctx context.Context, tagNames []string) ([]*proto.Post, error)
}

// Service implements IServer.
type Service struct {
	ctx context.Context

	// 用于主动关闭服务并重启。
	cancel func()

	// 计划重启。
	scheduledUpdate atomic.Bool

	home *url.URL

	// 服务器默认的时区。
	timeLocation *time.Location

	cfg *config.Config

	options utils.PluginStorage
	runtime *runtime_config.Runtime

	// 服务端渲染的时候一些模块会要求解析 URL（包含相对文件和绝对文件），
	// 所以需要用到主题相关的根文件系统。
	themeRootFS fs.FS
	postDataFS  theme_fs.FS
	// 动态管理的静态文件。
	userRoots *roots.Root

	db   *sql.DB
	tdb  *taorm.DB
	auth *auth.Auth
	mux  *http.ServeMux

	notifier      proto.NotifyServer
	cmtntf        *comment_notify.CommentNotifier
	cmtNotifyTask *_CommentNotificationTask
	cmtgeo        *commentgeo.Task

	// 通用内存缓存
	cache *lru.TTLCache[string, any]
	// 通用基于文件的缓存
	fileCache *cache.FileCache

	// 图片元数据缓存任务。
	exifTask *exif.Task
	// 朋友头像数据缓存任务
	friendsTask *friends.Task
	// 提醒事项任务
	remindersTask *reminders.Task
	// rss 任务
	rssTask *rss.Task

	// 文章内容缓存。
	// NOTE：缓存 Key 是跟文章编号和内容选项相关的（因为内容选项不同内容也就不同），
	// 所以存在一篇文章有多个缓存的问题。所以后面单独加了一个缓存来记录一篇文章对应了哪些
	// 与之相关的内容缓存。创建/删除内容缓存时，同步更新两个缓存。
	// TODO 我是不是应该直接缓存 *Post？不过好像也挺好改的。
	postContentCaches    *lru.TTLCache[_PostContentCacheKey, string]
	postCaches           *cache.RelativeCacheKeys[int64, _PostContentCacheKey]
	commentContentCaches *lru.TTLCache[_PostContentCacheKey, string]
	commentCaches        *cache.RelativeCacheKeys[int64, _PostContentCacheKey]

	avatarCache *cache.AvatarHash

	// 搜索引擎启动需要时间，所以如果网站一运行即搜索，则可能出现引擎不可用
	// 的情况，此时此值为空。
	searcher         atomic.Pointer[search.Engine]
	onceInitSearcher sync.Once

	maintenance *utils.Maintenance

	// 网站图标，临时放这儿。
	favicon *favicon.Favicon

	exporter *_Exporter
	// 证书剩余天数。
	// >= 0 表示值有效。
	certDaysLeft atomic.Int32
	// 域名有效期
	// >= 0 表示值有效。
	domainExpirationDaysLeft atomic.Int32

	proto.AuthServer
	proto.TaoBlogServer
	proto.ManagementServer
	proto.SearchServer
}

func (s *Service) Favicon() *favicon.Favicon {
	return s.favicon
}

func New(ctx context.Context, sr grpc.ServiceRegistrar, cfg *config.Config, db *sql.DB, rc *runtime_config.Runtime, auther *auth.Auth, mux *http.ServeMux, options ...With) *Service {
	s := &Service{
		ctx: ctx,

		notifier: &proto.UnimplementedNotifyServer{},

		cfg:         cfg,
		runtime:     rc,
		themeRootFS: embed.FS{},
		postDataFS:  &theme_fs.Empty{},

		// TODO 可配置使用的时区，而不是使用服务器当前时间或者硬编码成+8时区。
		timeLocation: time.Now().Location(),

		db:   db,
		tdb:  taorm.NewDB(db),
		auth: auther,
		mux:  mux,

		cache:     lru.NewTTLCache[string, any](10240),
		fileCache: cache.NewFileCache(ctx, ``),

		postContentCaches:    lru.NewTTLCache[_PostContentCacheKey, string](10240),
		postCaches:           cache.NewRelativeCacheKeys[int64, _PostContentCacheKey](),
		commentContentCaches: lru.NewTTLCache[_PostContentCacheKey, string](10240),
		commentCaches:        cache.NewRelativeCacheKeys[int64, _PostContentCacheKey](),

		maintenance: utils.NewMaintenance(),

		favicon: favicon.NewFavicon(),
	}

	for _, opt := range options {
		opt(s)
	}

	s.options = &_PluginStorage{
		ss:     s,
		prefix: ``,
	}

	s.userRoots = roots.New(s.GetPluginStorage(`roots`), s.mux)

	if u, err := url.Parse(cfg.Site.Home); err != nil {
		panic(err)
	} else {
		s.home = u
	}

	s.avatarCache = cache.NewAvatarHash()

	s.cmtntf = comment_notify.New(s.notifier)
	s.cmtNotifyTask = NewCommentNotificationTask(s, s.GetPluginStorage(`comment_notify`))
	s.cmtgeo = commentgeo.NewTask(s.GetPluginStorage(`cmt_geo`))

	s.cacheAllCommenterData()

	exifDebouncer := utils.NewBatchDebouncer(time.Second*10, func(id int) {
		s.deletePostContentCacheFor(int64(id))
		s.updatePostMetadataTime(int64(id), time.Now())
	})
	s.exifTask = exif.NewTask(s.GetPluginStorage(`exif`), func(id int) {
		exifDebouncer.Enter(id)
	})

	s.friendsTask = friends.NewTask(s.GetPluginStorage(`friends`), func(postID int) {
		s.deletePostContentCacheFor(int64(postID))
		s.updatePostMetadataTime(int64(postID), time.Now())
	})

	s.remindersTask = reminders.NewTask(ctx, s,
		func(id int) {
			s.deletePostContentCacheFor(int64(id))
			s.updatePostMetadataTime(int64(id), time.Now())
		},
		s.GetPluginStorage(`reminders`),
	)
	// TODO 注册到全局的，可能会导致测试冲突
	addons.Handle(`/reminders/`, http.StripPrefix(`/reminders`, s.remindersTask.CalenderService()))

	s.rssTask = rss.NewTask(ctx, s.GetPluginStorage(`rss`), func(pid int) {
		s.deletePostContentCacheFor(int64(pid))
		s.updatePostMetadataTime(int64(pid), time.Now())
	})

	s.certDaysLeft.Store(-1)
	s.domainExpirationDaysLeft.Store(-1)
	s.exporter = _NewExporter(s)

	if value, err := s.options.GetString(`favicon`); err == nil {
		d, _ := base64.RawURLEncoding.DecodeString(value)
		s.favicon.SetData(time.Now(), d)
	}

	proto.RegisterAuthServer(sr, s)
	proto.RegisterTaoBlogServer(sr, s)
	proto.RegisterManagementServer(sr, s)
	proto.RegisterSearchServer(sr, s)

	utilOptions := []UtilOption{}
	if ak := cfg.Others.Geo.Baidu.AccessKey; ak != `` {
		utilOptions = append(utilOptions, WithBaidu(ak, cfg.Site.Home))
	}
	utilsService := NewUtils(utilOptions...)
	proto.RegisterUtilsServer(sr, utilsService)

	return s
}

const noPerm = `此操作无权限。`

// 从 Context 中取出用户并且必须为 Admin/System，否则 panic。
func (s *Service) MustBeAdmin(ctx context.Context) *auth.AuthContext {
	return MustBeAdmin(ctx)
}

func (s *Service) MustCanCreatePost(ctx context.Context) *auth.AuthContext {
	ac := auth.Context(ctx)
	if ac == nil {
		panic("AuthContext 不应为 nil")
	}
	if ac.User.IsGuest() {
		panic(status.Error(codes.PermissionDenied, noPerm))
	}
	return ac
}

func MustBeAdmin(ctx context.Context) *auth.AuthContext {
	ac := auth.Context(ctx)
	if ac == nil {
		panic("AuthContext 不应为 nil")
	}
	if !ac.User.IsAdmin() && !ac.User.IsSystem() {
		panic(status.Error(codes.PermissionDenied, noPerm))
	}
	return ac
}

// Ping ...
func (s *Service) Ping(ctx context.Context, in *proto.PingRequest) (*proto.PingResponse, error) {
	return &proto.PingResponse{
		Pong: `pong`,
	}, nil
}

// Config ...
func (s *Service) Config() *config.Config {
	return s.cfg
}

func (s *Service) MustTxCall(callback func(s *Service) error) {
	if err := s.TxCall(callback); err != nil {
		panic(err)
	}
}

func (s *Service) MustTxCallNoError(callback func(s *Service)) {
	if err := s.TxCall(func(txs *Service) error {
		callback(txs)
		return nil
	}); err != nil {
		panic(err)
	}
}

// TODO 去掉，复制 s 非常有问题。
func (s *Service) TxCall(callback func(s *Service) error) error {
	if s.tdb.IsTx() {
		return callback(s)
	}
	return s.tdb.TxCall(func(tx *taorm.DB) error {
		txs := *s
		txs.tdb = tx
		txs.options = &_PluginStorage{
			ss:     &txs,
			prefix: ``,
		}
		return callback(&txs)
	})
}

// 是否在维护模式。
// 1. 手动进入。
// 2. 自动升级过程中。
// https://github.com/movsb/taoblog/commit/c4428d7
func (s *Service) MaintenanceMode() *utils.Maintenance {
	return s.maintenance
}

func (s *Service) GetInfo(ctx context.Context, in *proto.GetInfoRequest) (*proto.GetInfoResponse, error) {
	t := time.Now()
	if modified := s.GetDefaultIntegerOption("last_post_time", 0); modified > 0 {
		t = time.Unix(modified, 0)
	}

	out := &proto.GetInfoResponse{
		Name:         s.cfg.Site.Name,
		Description:  s.cfg.Site.Description,
		Home:         strings.TrimSuffix(s.cfg.Site.Home, "/"),
		Commit:       version.GitCommit,
		LastPostedAt: int32(t.Unix()),

		CertDaysLeft:   s.certDaysLeft.Load(),
		DomainDaysLeft: s.domainExpirationDaysLeft.Load(),

		ScheduledUpdate: s.scheduledUpdate.Load(),
	}

	return out, nil
}

func (s *Service) GetCurrentTimezone() *time.Location {
	return s.timeLocation
}

func (s *Service) SetCertDays(n int) {
	s.certDaysLeft.Store(int32(n))
	s.exporter.certDaysLeft.Set(float64(n))
}
func (s *Service) SetDomainDays(n int) {
	s.domainExpirationDaysLeft.Store(int32(n))
	s.exporter.domainDaysLeft.Set(float64(n))
}
