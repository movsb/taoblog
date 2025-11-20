package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/movsb/taoblog/cmd/config"
	"github.com/movsb/taoblog/gateway/handlers/favicon"
	"github.com/movsb/taoblog/gateway/handlers/roots"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/crypto"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/modules/version"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/cache"
	"github.com/movsb/taoblog/service/modules/calendar"
	"github.com/movsb/taoblog/service/modules/calendar/solar"
	commentgeo "github.com/movsb/taoblog/service/modules/comment_geo"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taoblog/service/modules/renderers/blur_image"
	"github.com/movsb/taoblog/service/modules/renderers/exif"
	"github.com/movsb/taoblog/service/modules/renderers/friends"
	"github.com/movsb/taoblog/service/modules/renderers/genealogy"
	"github.com/movsb/taoblog/service/modules/renderers/reminders"
	runtime_config "github.com/movsb/taoblog/service/modules/runtime"
	"github.com/movsb/taoblog/service/modules/search"
	"github.com/movsb/taoblog/service/modules/storage"
	"github.com/movsb/taoblog/setup/migration"
	theme_fs "github.com/movsb/taoblog/theme/modules/fs"
	"github.com/movsb/taoblog/theme/styling"
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

	// 配置文件中写的站点地址。
	// 可以被测试环境修改成临时地址。
	// 对外的链接生成时使用这个地址。
	getHome func() string

	// 服务器默认的时区。
	timeLocation *time.Location

	cfg *config.Config

	options utils.PluginStorage
	runtime *runtime_config.Runtime

	// TODO 包装并鉴权。
	postDataFS  *WrappedPostFilesFileSystem
	mainStorage *storage.SQLite

	// 动态管理的静态文件。
	userRoots *roots.Root

	// 已经注册的外部文件存储。
	fileURLGetters sync.Map
	// 防止每个请求总是生成不同的 URL。
	fileURLs *lru.LRUCache[_FileURLCacheKey, *_FileURLCacheValue]

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
	// rssTask *rss.Task
	// 模糊缩略图任务
	thumbHashTask *blur_image.Task
	// 族谱图任务
	genealogyTask *genealogy.Task

	calendar *calendar.CalenderService
	aesGCM   *crypto.AesGcm

	// 文章内容缓存。
	// NOTE：缓存 Key 是跟文章编号和内容选项相关的（因为内容选项不同内容也就不同），
	// 所以存在一篇文章有多个缓存的问题。所以后面单独加了一个缓存来记录一篇文章对应了哪些
	// 与之相关的内容缓存。创建/删除内容缓存时，同步更新两个缓存。
	// TODO 我是不是应该直接缓存 *Post？不过好像也挺好改的。
	postFullCaches       *lru.TTLCache[int64, *models.Post]
	postContentCaches    *lru.TTLCache[_PostContentCacheKey, string]
	postCaches           *cache.RelativeCacheKeys[int64, _PostContentCacheKey]
	commentContentCaches *lru.TTLCache[_PostContentCacheKey, string]
	commentCaches        *cache.RelativeCacheKeys[int64, _PostContentCacheKey]

	// 神经病一样，没有 clear 方法，要清空只能新建一个。
	relatesCaches atomic.Pointer[lru.TTLCache[int64, []*proto.Post]]

	avatarCache *cache.AvatarHash

	// 搜索引擎启动需要时间，所以如果网站一运行即搜索，则可能出现引擎不可用
	// 的情况，此时此值为空。
	searcher         atomic.Pointer[search.Engine]
	onceInitSearcher sync.Once

	maintenance *utils.Maintenance

	// 网站图标，临时放这儿。
	favicon *favicon.Favicon

	// 证书剩余天数。
	// >= 0 表示值有效。
	certDaysLeft atomic.Int32
	// 域名有效期
	// >= 0 表示值有效。
	domainExpirationDaysLeft atomic.Int32

	// 最后同步与备份时间。
	lastBackupAt atomic.Int32
	lastSyncAt   atomic.Int32

	// 存储状态报告。
	postsStorageSize atomic.Int64
	filesStorageSize atomic.Int64

	proto.AuthServer
	proto.TaoBlogServer
	proto.ManagementServer
	proto.SearchServer
}

func (s *Service) Favicon() *favicon.Favicon {
	return s.favicon
}

func New(ctx context.Context, sr grpc.ServiceRegistrar, cfg *config.Config, db *taorm.DB, rc *runtime_config.Runtime, auther *auth.Auth, mux *http.ServeMux, options ...With) *Service {
	s := &Service{
		ctx: ctx,

		getHome: cfg.Site.GetHome,

		AuthServer: auther.Passkeys(),

		notifier: &proto.UnimplementedNotifyServer{},

		cfg:        cfg,
		runtime:    rc,
		postDataFS: &WrappedPostFilesFileSystem{FS: &theme_fs.Empty{}},

		// TODO 可配置使用的时区，而不是使用服务器当前时间或者硬编码成+8时区。
		timeLocation: time.Now().Location(),

		db:   db.Underlying(),
		tdb:  db,
		auth: auther,
		mux:  mux,

		cache:     lru.NewTTLCache[string, any](10240),
		fileCache: cache.NewFileCache(ctx, taorm.NewDB(migration.InitCache(``))),
		fileURLs:  lru.NewLRUCache[_FileURLCacheKey, *_FileURLCacheValue](1024),

		postFullCaches:       lru.NewTTLCache[int64, *models.Post](1024),
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

	s.mustInitCrypto()

	s.calendar = calendar.NewCalendarService(time.Now)
	mux.HandleFunc(`GET /v3/calendars`, s.handleGetCalendar)

	s.userRoots = roots.New(s.GetPluginStorage(`roots`), s.mux)
	s.avatarCache = cache.NewAvatarHash()

	s.cmtntf = comment_notify.New(s.notifier)
	s.cmtNotifyTask = NewCommentNotificationTask(s, s.GetPluginStorage(`comment_notify`))
	s.cmtgeo = commentgeo.NewTask(s.GetPluginStorage(`cmt_geo`))

	s.cacheAllCommenterData()

	exifDebouncer := utils.NewBatchDebouncer(time.Second*10, func(id int) {
		s.deletePostContentCacheFor(int64(id))
		s.updatePostMetadataTime(int64(id), time.Now())
	})
	s.exifTask = exif.NewTask(s.fileCache, exifDebouncer.Enter)

	s.friendsTask = friends.NewTask(s.fileCache, func(postID int) {
		s.deletePostContentCacheFor(int64(postID))
		s.updatePostMetadataTime(int64(postID), time.Now())
	})

	s.remindersTask = reminders.NewTask(ctx, s,
		func(id int) {
			s.deletePostContentCacheFor(int64(id))
			s.updatePostMetadataTime(int64(id), time.Now())
		},
		s.GetPluginStorage(`reminders`),
		s.calendar,
		cfg.Site.GetTimezoneLocation,
	)

	s.genealogyTask = genealogy.NewTask(ctx, s,
		s.GetPluginStorage(`genealogy`),
		s.calendar,
	)

	s.thumbHashTask = blur_image.NewTask(s.ctx, s.GetPluginStorage(`thumb_hash`), s.mainStorage, func(pid int) {
		s.deletePostContentCacheFor(int64(pid))
		s.updatePostMetadataTime(int64(pid), time.Now())
	})

	// s.rssTask = rss.NewTask(ctx, s.GetPluginStorage(`rss`), func(pid int) {
	// 	s.deletePostContentCacheFor(int64(pid))
	// 	s.updatePostMetadataTime(int64(pid), time.Now())
	// })

	s.relatesCaches.Store(lru.NewTTLCache[int64, []*proto.Post](128))

	s.certDaysLeft.Store(-1)
	s.domainExpirationDaysLeft.Store(-1)

	if value, err := s.options.GetString(`favicon`); err == nil {
		u, _ := utils.ParseDataURL(value)
		if u != nil {
			s.favicon.SetData(time.Now(), u)
		}
	}
	if value, err := s.options.GetInteger(`styling_page_id`); err == nil {
		s.postDataFS.Register(int(value), styling.Root)
	}

	proto.RegisterAuthServer(sr, s)
	proto.RegisterTaoBlogServer(sr, s)
	proto.RegisterManagementServer(sr, s)
	proto.RegisterSearchServer(sr, s)

	utilOptions := []UtilOption{}
	if ak := cfg.Others.Geo.Baidu.AccessKey; ak != `` {
		utilOptions = append(utilOptions, WithBaidu(ak, s.getHome))
		utilOptions = append(utilOptions, WithTimezone(cfg.Site.GetTimezoneLocation))
	}
	utilsService := NewUtils(utilOptions...)
	proto.RegisterUtilsServer(sr, utilsService)

	return s
}

func (s *Service) mustInitCrypto() {
	aesKey, err := s.options.GetString(`aes_key`)
	if taorm.IsNotFoundError(err) {
		key := crypto.NewSecret()
		utils.Must(s.options.SetString(`aes_key`, key.String()))
		log.Println(`加密密钥：`, key.String())
		aesKey = key.String()
	}
	s.aesGCM = utils.Must1(crypto.NewAesGcm(utils.Must1(crypto.SecretFromString(aesKey))))
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
		Name:        s.cfg.Site.GetName(),
		Description: s.cfg.Site.GetDescription(),
		Home:        s.getHome(),
		Commit:      version.GitCommit,
		Uptime:      int32(time.Since(version.Time).Seconds()),

		CertDaysLeft:   s.certDaysLeft.Load(),
		DomainDaysLeft: s.domainExpirationDaysLeft.Load(),

		LastPostedAt: int32(t.Unix()),
		LastBackupAt: s.lastBackupAt.Load(),
		LastSyncAt:   s.lastSyncAt.Load(),

		ScheduledUpdate: s.scheduledUpdate.Load(),

		Storage: &proto.GetInfoResponse_StorageStatus{
			Posts: s.postsStorageSize.Load(),
			Files: s.filesStorageSize.Load(),
		},
	}

	return out, nil
}

func (s *Service) GetCurrentTimezone() *time.Location {
	return s.timeLocation
}

// 可能小于0
func (s *Service) SetCertDays(n int) {
	s.certDaysLeft.Store(int32(n))

	s.calendar.Remove(svcCalKind, func(e *calendar.Event) bool {
		uuid, _ := e.Tags[`uuid`]
		return uuid == `cert_days`
	})

	if n >= 0 && n < 15 {
		st, et := solar.AllDay(time.Now())
		s.calendar.AddEvent(svcCalKind, &calendar.Event{
			Message: fmt.Sprintf(`证书剩余 %d 天`, n),
			Start:   st,
			End:     et,
			UserID:  auth.SystemID,
			PostID:  0,
			Tags: map[string]any{
				`uuid`: `cert_days`,
			},
		})
	}
}

// 可能小于0
func (s *Service) SetDomainDays(n int) {
	s.domainExpirationDaysLeft.Store(int32(n))

	s.calendar.Remove(svcCalKind, func(e *calendar.Event) bool {
		uuid, _ := e.Tags[`uuid`]
		return uuid == `domain_days`
	})

	if n >= 0 && n < 15 {
		st, et := solar.AllDay(time.Now())
		s.calendar.AddEvent(svcCalKind, &calendar.Event{
			Message: fmt.Sprintf(`域名剩余 %d 天`, n),
			Start:   st,
			End:     et,
			UserID:  auth.SystemID,
			PostID:  0,
			Tags: map[string]any{
				`uuid`: `domain_days`,
			},
		})
	}
}

func (s *Service) SetLastBackupAt(t time.Time) {
	s.lastBackupAt.Store(int32(t.Unix()))
}
func (s *Service) SetLastSyncAt(t time.Time) {
	s.lastSyncAt.Store(int32(t.Unix()))
}
func (s *Service) SetPostsStorageSize(size int64) {
	s.postsStorageSize.Store(size)
}
func (s *Service) SetFilesStorageSize(size int64) {
	s.filesStorageSize.Store(size)
}
