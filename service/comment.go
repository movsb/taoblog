package service

import (
	"bytes"
	"cmp"
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/modules/globals"
	"github.com/movsb/taoblog/modules/utils"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/comment_notify"
	"github.com/movsb/taorm"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func (s *Service) getCommentContentCached(ctx context.Context, id int64, sourceType, source string, postID int64, co *proto.PostContentOptions) (string, error) {
	ac := auth.Context(ctx)
	key := _PostContentCacheKey{
		ID:      id,
		Options: co.String(),
		UserID:  int(ac.User.ID),
		Public:  s.isPostPublic(ctx, int(postID)),
	}
	content, err, _ := s.commentContentCaches.GetOrLoad(ctx, key,
		func(ctx context.Context, key _PostContentCacheKey) (string, time.Duration, error) {
			// NOTE：带缓存的，默认认识总是安全的
			// TODO：用户的评论不应该渲染 #hashtags，好像没有意义。
			content, err := s.renderMarkdown(ctx, true, postID, id, sourceType, source, models.PostMeta{}, co, key.Public)
			if err != nil {
				return ``, 0, err
			}
			s.commentCaches.Append(id, key)
			return content, time.Hour, nil
		},
	)
	if err != nil {
		return ``, err
	}
	return content, nil
}

func (s *Service) deleteCommentContentCacheFor(id int64) {
	log.Println(`即将删除评论缓存：`, id)
	s.commentCaches.Delete(id, func(second _PostContentCacheKey) {
		s.commentContentCaches.Delete(second)
		log.Println(`删除评论缓存：`, second)
	})
}

func (s *Service) UserAvatarEphemeral(ctx context.Context, uid int, email string) uint32 {
	if uid > 0 {
		user, err := s.auth.GetUserByID(ctx, int64(uid))
		if err == nil {
			if len(user.Avatar.Data) > 0 {
				return s.avatarCache.User(int(user.ID))
			}
			// 如果用记有邮箱，始终使用用户邮箱，而不是评论的邮箱。
			if user.Email != `` {
				return s.avatarCache.Email(user.Email)
			}
		}
	}

	return s.avatarCache.Email(email)
}

func (s *Service) setCommentExtraFields(ctx context.Context, co *proto.PostContentOptions) func(c *proto.Comment) {
	ac := auth.Context(ctx)

	if co == nil {
		co = &proto.PostContentOptions{}
	}

	return func(c *proto.Comment) {
		c.Avatar = s.UserAvatarEphemeral(ctx, int(c.UserId), c.Email)

		// （同 IP 用户 & 5️⃣分钟内） 可编辑。
		// TODO: IP：并不严格判断，比如网吧、办公室可能具有相同 IP。所以限制了时间范围。
		// NOTE：管理员总是可以编辑，跟此值无关。
		// 只允许编辑 Markdown 评论。
		// TODO：其实也允许/也已经支持编辑早期的 HTML 评论，但是在保存的时候已经被转换成 Markdown。
		c.CanEdit = c.SourceType == `markdown` && (ac.RemoteAddr.String() == c.Ip && in5min(c.Date))

		if ac.User.IsGuest() {
			c.Email = ""
			c.Ip = ""
		} else {
			c.GeoLocation = s.cmtgeo.Get(c.Ip)
		}

		if co.WithContent {
			content, err := s.getCommentContentCached(ctx, c.Id, c.SourceType, c.Source, c.PostId, co)
			if err != nil {
				slog.Error("转换评论时出错：", slog.String(`error`, err.Error()))
				// 也不能干啥……
				// fallthrough
			}
			c.Content = content
		}
	}
}

// Deprecated.
func (s *Service) comments() *taorm.Stmt {
	return s.tdb.Model(models.Comment{})
}

// GetComment2 ...
func (s *Service) getComment2(name int64) *models.Comment {
	var comment models.Comment
	s.comments().Where("id=?", name).MustFind(&comment)
	return &comment
}

// 像二狗说的那样，服务启动时缓存所有的头像哈希值，
// 否则缓存的页面图片在服务重启后刷新时会加载失败。
// https://qwq.me/p/249/1#comment-506
// NOTE: ORM 不支持 distinct，所以没写。
func (s *Service) cacheAllCommenterData() {
	var comments models.Comments
	s.tdb.Select(`email,ip`).OrderBy(`date desc`).MustFind(&comments)
	dup := make(map[string]struct{})
	for _, c := range comments {
		if _, ok := dup[c.Email]; !ok {
			dup[c.Email] = struct{}{}
			_ = s.avatarCache.Email(c.Email)
		}
		if _, ok := dup[c.IP]; !ok {
			dup[c.IP] = struct{}{}
			s.cmtgeo.Get(c.IP)
		}
	}
}

func in5min(t int32) bool {
	return time.Since(time.Unix(int64(t), 0)) < time.Minute*5
}

// 更新评论。
//
// NOTE：只支持更新评论内容。
// NOTE：带上时间戳，防止异地多次更新的冲突（太严格了吧！）
// NOTE：带节流。
func (s *Service) UpdateComment(ctx context.Context, req *proto.UpdateCommentRequest) (*proto.Comment, error) {
	ac := auth.Context(ctx)
	cmtOld := s.getComment2(req.Comment.Id)

	// 1. 同 IP 的游客五分钟内允许
	// 2. 管理员允许
	// 3. 其它登录用户五分钟内
	allowEdit := (ac.User.IsGuest() && ac.RemoteAddr.String() == cmtOld.IP && in5min(cmtOld.Date)) ||
		ac.User.IsAdmin() ||
		!ac.User.IsGuest() && in5min(cmtOld.Date)
	if !allowEdit {
		return nil, status.Error(codes.PermissionDenied, `超时或无权限编辑评论`)
	}

	// 其他已登录用户只能改自己创建的
	if !ac.User.IsGuest() && !ac.User.IsAdmin() && cmtOld.UserID > 0 && ac.User.ID != int64(cmtOld.UserID) {
		return nil, status.Error(codes.PermissionDenied, `不能编辑非由自己创建的评论`)
	}

	var comment models.Comment

	if req.Comment != nil && req.UpdateMask != nil && req.UpdateMask.Paths != nil {
		data := map[string]any{}
		var hasSourceType, hasSource bool
		var hasModified bool
		for _, mask := range req.UpdateMask.Paths {
			switch mask {
			default:
				panic(`unknown update_mask field`)
			case `source_type`:
				hasSourceType = true
				data[`source_type`] = req.Comment.SourceType
			case `source`:
				hasSource = true
				data[`source`] = req.Comment.Source
			case `modified`:
				hasModified = true
				data[`modified`] = time.Now().Unix()
			case `modified_timezone`:
				// hasModified = true
				data[mask] = req.Comment.ModifiedTimezone
			}
		}
		if !hasModified {
			return nil, status.Error(codes.Aborted, `更新评论需要带上评论本身的修改时间。`)
		} else if cmtOld.Modified != req.Comment.Modified {
			return nil, status.Error(codes.Aborted, `当前评论内容已在其它地方被修改过，请刷新后重新提交。`)
		}
		if hasSourceType != hasSource {
			panic(`source_type and source must be both specified`)
		}
		if hasSourceType {
			if strings.TrimSpace(req.Comment.Source) == "" {
				return nil, status.Error(codes.InvalidArgument, `评论内容不能为空。`)
			}
			if req.Comment.SourceType != `markdown` {
				return nil, status.Error(codes.InvalidArgument, `不再允许发表非 Markdown 评论。`)
			}
			// NOTE：管理员如果修改用户评论后如果带 HTML，则用户无法再提交保存。
			// 是不是应该限制下呢？
			if _, err := s.renderMarkdown(ctx, ac.User.IsAdmin(), cmtOld.PostID, comment.ID, req.Comment.SourceType, req.Comment.Source, models.PostMeta{}, co.For(co.UpdateCommentCheck), s.isPostPublic(ctx, int(cmtOld.PostID))); err != nil {
				return nil, err
			}
		}
		s.MustTxCall(func(txs *Service) error {
			txs.tdb.Model(models.Comment{}).Where(`id=?`, req.Comment.Id).MustUpdateMap(data)
			txs.tdb.Where(`id=?`, req.Comment.Id).MustFind(&comment)
			txs.updatePostCommentCount(comment.PostID, time.Now())
			txs.deleteCommentContentCacheFor(comment.ID)
			return nil
		})
	} else {
		s.tdb.Where(`id=?`, req.Comment.Id).MustFind(&comment)
	}

	return comment.ToProto(s.setCommentExtraFields(ctx, co.For(co.UpdateCommentReturn))), nil
}

// 删除评论。
//
// 会递归地删除其所有子评论。
//
// TODO：清理数据库的脏数据（有一部分 parent 评论已经不存在）。
func (s *Service) DeleteComment(ctx context.Context, in *proto.DeleteCommentRequest) (*proto.DeleteCommentResponse, error) {
	s.MustBeAdmin(ctx)

	// 点击“删除”的评论。
	cmt := s.getComment2(int64(in.Id))

	// 此评论顶级评论下的所有评论
	var all models.Comments
	if cmt.Root == 0 {
		s.tdb.Where(`root=?`, cmt.ID).MustFind(&all)
	} else {
		s.tdb.Where(`root=?`, cmt.Root).MustFind(&all)
	}

	// 找出所有以待删除评论为 “root” 的评论。
	toDelete := []int64{int64(in.Id)}
	for i := 0; i < len(toDelete); i++ {
		for j := 0; j < len(all); j++ {
			if all[j].Parent == toDelete[i] {
				toDelete = append(toDelete, all[j].ID)
			}
		}
	}

	s.comments().Where("id IN (?)", toDelete).MustDelete()

	s.updatePostCommentCount(cmt.PostID, time.Now())
	s.updateCommentsCount()

	for _, d := range toDelete {
		s.deleteCommentContentCacheFor(d)
	}

	return &proto.DeleteCommentResponse{}, nil
}

func (s *Service) ListComments(ctx context.Context, in *proto.ListCommentsRequest) (*proto.ListCommentsResponse, error) {
	ac := auth.Context(ctx)

	if in.Limit <= 0 || in.Limit > 100 {
		panic(status.Errorf(codes.InvalidArgument, `limit out of range`))
	}

	// TODO ensure that fields must include root etc to be used later.
	// TODO verify fields that are sanitized.
	stmt := s.tdb.Select(`distinct comments.*`)
	stmt.WhereIf(in.PostId > 0, "post_id=?", in.PostId)
	stmt.Limit(in.Limit).Offset(in.Offset).OrderBy(in.OrderBy)
	if ac.User.IsGuest() {
		stmt.InnerJoin("posts", "comments.post_id = posts.id AND posts.status = 'public'")
	} else if ac.User.IsSystem() {
		// nothing to do
	} else {
		switch in.Ownership {
		default:
			return nil, fmt.Errorf(`未知所有者。`)
		case proto.Ownership_OwnershipMine:
			stmt.InnerJoin(models.Post{}, `comments.post_id=posts.id`)
			stmt.Where(`posts.user_id=?`, ac.User.ID)
		case proto.Ownership_OwnershipTheir:
			stmt.InnerJoin(models.Post{}, `comments.post_id=posts.id`)
			stmt.InnerJoin(models.AccessControlEntry{}, `comments.post_id=acl.post_id`)
			stmt.Where(
				`posts.user_id!=? AND (posts.status=? OR (acl.user_id=? AND posts.status = ?))`,
				ac.User.ID, models.PostStatusPublic, ac.User.ID, models.PostStatusPartial,
			)
		case proto.Ownership_OwnershipMineAndShared:
			stmt.InnerJoin(models.Post{}, `comments.post_id=posts.id`)
			stmt.LeftJoin(models.AccessControlEntry{}, `comments.post_id=acl.post_id`)
			stmt.Where(
				`posts.user_id=? OR (acl.user_id=? AND posts.status = ?)`,
				ac.User.ID, ac.User.ID, models.PostStatusPartial,
			)
		case proto.Ownership_OwnershipUnknown, proto.Ownership_OwnershipAll:
			stmt.InnerJoin(models.Post{}, `comments.post_id=posts.id`)
			stmt.LeftJoin(models.AccessControlEntry{}, `posts.id = acl.post_id`)
			stmt.Where(
				`posts.user_id=? OR (posts.status=? OR (acl.user_id=? AND posts.status = ?))`,
				ac.User.ID, models.PostStatusPublic, ac.User.ID, models.PostStatusPartial,
			)
		}
	}
	if len(in.Types) > 0 {
		stmt.InnerJoin(`posts`, `comments.post_id = posts.id AND posts.type in (?)`, in.Types)
	}

	var comments models.Comments
	stmt.MustFind(&comments)

	protoComments := comments.ToProto(s.setCommentExtraFields(ctx, in.ContentOptions))
	return &proto.ListCommentsResponse{Comments: protoComments}, nil
}

// 由 GetPostComments 改写而来，私有函数，不鉴权。
// TODO 缓存。
func (s *Service) getPostComments(ctx context.Context, postID int64) ([]*proto.Comment, error) {
	var comments models.Comments
	s.tdb.Where("post_id=?", postID).MustFind(&comments)
	return comments.ToProto(s.setCommentExtraFields(ctx, co.For(co.GetPostComments))), nil
}

func (s *Service) GetAllCommentsCount() int64 {
	var count int64
	s.tdb.Model(models.Comment{}).Select("count(1) as count").Find(&count)
	return count
}

const (
	maxNicknameLen = 32
	maxEmailLen    = 64
	maxUrlLen      = 256
	maxContentLen  = 4096
)

// 创建一条评论。
//
// 前期验证项：
//
// Author 不为空、长度不超限
// Email 不为空、长度不超限
// URL 可为空，若不为空，需要为正确的 URL，可不加 http 前缀，自动加
// SourceType 只能为 markdown
// Source 不为空且不超限。
//
// 检查昵称是否被允许
// 检查邮箱是否被允许
//
// 检查项：
// post id 存在
// parent 是否存在
// 且和 parent 的 post id 一样
// root 置为 parent / parent的root
// IP 从请求中自动获取，忽略传入。
// Date 服务端的当前时间戳，忽略传入。
// Content 自动由 source 生成。
//
// NOTE: 默认的 modified 修改时间为 0，表示从未被修改过。
// NOTE: 带节流。
func (s *Service) CreateComment(ctx context.Context, in *proto.Comment) (*proto.Comment, error) {
	ac := auth.Context(ctx)

	// 尽早查询地理信息
	s.cmtgeo.Get(ac.RemoteAddr.String())

	now := time.Now()

	c := models.Comment{
		PostID:     in.PostId,
		Parent:     in.Parent,
		Author:     strings.TrimSpace(in.Author),
		Email:      strings.TrimSpace(in.Email),
		UserID:     int32(ac.User.ID),
		URL:        strings.TrimSpace(in.Url),
		IP:         ac.RemoteAddr.String(),
		Date:       0,
		SourceType: in.SourceType,
		Source:     in.Source,
	}

	if ac.User.ID > 0 {
		c.Author = ac.User.Nickname
		if ac.User.Email != `` {
			c.Email = ac.User.Email
		}
		c.URL = ``
	}

	c.ModifiedTimezone = in.ModifiedTimezone
	if in.Modified > 0 {
		c.Modified = in.Modified
	}

	c.DateTimezone = in.DateTimezone
	if c.ModifiedTimezone == `` && c.DateTimezone != `` {
		c.ModifiedTimezone = c.DateTimezone
	}

	if in.Date > 0 {
		c.Date = in.Date
		if in.Modified == 0 {
			c.Modified = c.Date
			c.ModifiedTimezone = c.DateTimezone
		}
	} else {
		c.Date = int32(now.Unix())
		c.Modified = int32(now.Unix())
	}

	if c.Author == "" {
		return nil, status.Error(codes.InvalidArgument, `昵称不能为空`)
	}
	if utf8.RuneCountInString(c.Author) >= maxNicknameLen {
		return nil, status.Errorf(codes.InvalidArgument, `昵称太长（最长 %d 个字符）`, maxNicknameLen)
	}

	if !utils.IsEmail(c.Email) {
		return nil, status.Errorf(codes.InvalidArgument, `邮箱格式不正确`)
	}
	if utf8.RuneCountInString(c.Email) >= maxEmailLen {
		return nil, status.Errorf(codes.InvalidArgument, `邮箱太长（最长 %d 个字符）`, maxEmailLen)
	}

	if c.URL != "" {
		if !strings.Contains(c.URL, "://") {
			c.URL = `http://` + c.URL
		}
		if !utils.IsURL(c.URL, false) {
			return nil, status.Errorf(codes.InvalidArgument, `网址格式不正确`)
		}
	}
	if utf8.RuneCountInString(c.URL) >= maxUrlLen {
		return nil, status.Errorf(codes.InvalidArgument, `网址太长（最长 %d 个字符）`, maxUrlLen)
	}

	if c.SourceType != `markdown` && !ac.User.IsAdmin() {
		return nil, status.Error(codes.InvalidArgument, `不再允许发表非 Markdown 评论。`)
	}
	if strings.TrimSpace(c.Source) == "" {
		return nil, status.Error(codes.InvalidArgument, `评论内容不能为空`)
	}
	if utf8.RuneCountInString(c.Source) >= maxContentLen {
		return nil, status.Error(codes.InvalidArgument, `评论内容太长`)
	}

	if ac.User.IsGuest() {
		if c.Author != "" && strings.Contains(in.Author, "作者") {
			return nil, status.Error(codes.InvalidArgument, "昵称中不应包含“作者”两字")
		}
	}

	if c.SourceType == `markdown` {
		if _, err := s.renderMarkdown(ctx, ac.User.IsAdmin(), c.PostID, 0, c.SourceType, c.Source, models.PostMeta{}, co.For(co.CreateCommentCheck), s.isPostPublic(ctx, int(in.PostId))); err != nil {
			return nil, err
		}
	}

	s.MustTxCall(func(txs *Service) error {
		if c.Parent > 0 {
			pc := txs.getComment2(c.Parent)
			if pc.Root != 0 {
				c.Root = pc.Root
			} else {
				c.Root = pc.ID
			}
			if c.PostID != pc.PostID {
				panic(status.Error(codes.InvalidArgument, `不是同一篇文章的父评论。`))
			}
		} else {
			c.Root = 0
		}
		txs.tdb.Model(&c).MustCreate()
		txs.updatePostCommentCount(c.PostID, time.Unix(int64(c.Date), 0))
		txs.updateCommentsCount()

		// 创建评论会改变评论条数在 tweets 页面的显示，所以有必要影响缓存。
		// 但是更新评论不会影响条数，目前进入页面一定会作 304 检测，所以评论列表一定会是最新的。
		txs.updateLastPostTime(time.Now())
		return nil
	})

	return c.ToProto(s.setCommentExtraFields(ctx, co.For(co.CreateCommentReturn))), nil
}

func (s *Service) updateCommentsCount() {
	count := s.GetAllCommentsCount()
	s.options.SetInteger("comment_count", count)
}

// SetCommentPostID 把某条顶级评论及其子评论转移到另一篇文章下
// TODO：禁止转移内容中引用了当前文章资源的评论，或者处理这个问题。
func (s *Service) SetCommentPostID(ctx context.Context, in *proto.SetCommentPostIDRequest) (*proto.SetCommentPostIDResponse, error) {
	s.MustBeAdmin(ctx)

	s.MustTxCall(func(txs *Service) error {
		cmt := txs.getComment2(in.Id)
		if cmt.Root != 0 {
			panic(`不能转移子评论`)
		}
		// 只是为了判断存在性。
		_, err := txs.GetPost(ctx, &proto.GetPostRequest{Id: int32(in.PostId)})
		if err != nil {
			return err
		}
		if cmt.PostID == in.PostId {
			panic(`不能转移到相同的文章`)
		}
		txs.tdb.From(cmt).
			Where(`post_id=?`, cmt.PostID).
			Where(`id=? OR root=?`, cmt.ID, cmt.ID).
			MustUpdateMap(map[string]any{
				`post_id`: in.PostId,
			})
		txs.updatePostCommentCount(cmt.PostID, time.Now())
		txs.updatePostCommentCount(in.PostId, time.Now())
		log.Printf("Transferred comments %d to post %d", cmt.ID, in.PostId)
		return nil
	})

	return &proto.SetCommentPostIDResponse{}, nil
}

func (s *Service) PreviewComment(ctx context.Context, in *proto.PreviewCommentRequest) (*proto.PreviewCommentResponse, error) {
	ac := auth.Context(ctx)
	content, err := s.renderMarkdown(ctx, ac.User.IsAdmin(), int64(in.PostId), 0, `markdown`, in.Markdown, models.PostMeta{}, co.For(co.PreviewComment), s.isPostPublic(ctx, int(in.PostId)))
	return &proto.PreviewCommentResponse{Html: content}, err
}

type _CommentNotificationTask struct {
	s       *Service
	storage utils.PluginStorage

	// 扫描新评论的时间间隔。
	// 不能低于一秒（因为有个成功 sleep 1s）。
	scanInterval time.Duration
	// 新评论产生后需要延迟多久才被扫描到（方便新发表后的编辑操作）。
	dateDelay time.Duration
}

// 记录上一次的评论编号，以跳过同一秒内可能重复的评论。
// NOTE: 假设评论的编号是递增的，否则会失败。
func NewCommentNotificationTask(s *Service, storage utils.PluginStorage) *_CommentNotificationTask {
	t := _CommentNotificationTask{
		s:       s,
		storage: storage,

		scanInterval: time.Second * 15,
		dateDelay:    time.Minute * 1,
	}
	if _, err := t.storage.GetInteger(`id`); err != nil {
		var max int64
		if err := t.s.tdb.Raw(`SELECT seq FROM sqlite_sequence WHERE name=?`, models.Comment{}.TableName()).Find(&max); err != nil {
			if taorm.IsNotFoundError(err) {
				max = 0
			} else {
				panic(err)
			}
		}
		if err := t.storage.SetInteger(`id`, max); err != nil {
			panic(err)
		}
		log.Println(`当前评论的最大编号：`, max)
	}
	go t.Run(s.ctx)
	return &t
}

func (t *_CommentNotificationTask) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println(`退出评论通知任务`)
			return
		default:
		}
		if err := t.runOnce(ctx); err == nil {
			time.Sleep(time.Second)
		} else {
			if !taorm.IsNotFoundError(err) {
				log.Println(err)
			}
			time.Sleep(t.scanInterval)
		}
	}
}

func (t *_CommentNotificationTask) runOnce(ctx context.Context) error {
	c, err := t.getNewComment()
	if err != nil {
		return err
	}
	log.Println(`找到新评论：`, c)
	if err := t.queueForSingle(c); err != nil {
		return err
	}
	if err := t.storage.SetInteger(`id`, c.ID); err != nil {
		return err
	}
	return nil
}

func (t *_CommentNotificationTask) getLast() (int64, error) {
	return t.storage.GetInteger(`id`)
}

// 取得一条自上次检查以来新产生的评论。
func (t *_CommentNotificationTask) getNewComment() (*models.Comment, error) {
	lastId, err := t.getLast()
	if err != nil {
		return nil, err
	}

	dateBefore := time.Now().Add(-t.dateDelay)

	var c models.Comment
	if err := t.s.tdb.Where(`id > ? AND date < ?`, lastId, dateBefore.Unix()).OrderBy(`id ASC`).Find(&c); err != nil {
		return nil, err
	}

	// 正常来说扫描出的评论时间不可能超过 dateDelay，
	// 否则属于错误扫描（比如意外扫描到 ID 为 1 的时的评论。
	// TODO 检查评论时间。

	return &c, nil
}

func (t *_CommentNotificationTask) queueForSingle(c *models.Comment) error {
	post, err := t.s.GetPost(auth.SystemForLocal(context.Background()), &proto.GetPostRequest{
		Id: int32(c.PostID),
		GetPostOptions: &proto.GetPostOptions{
			WithLink:       proto.LinkKind_LinkKindFull,
			ContentOptions: co.For(co.CreateCommentGetPost),
		},
	})
	if err != nil {
		return err
	}

	pu, err := t.s.auth.GetUserByID(context.TODO(), int64(post.UserId))
	if err != nil {
		log.Println(`获取文章作者失败：`, err)
		return err
	}

	loc := globals.LoadTimezoneOrDefault(c.DateTimezone, time.Local)

	// 始终通知站长。
	data := comment_notify.AdminData{
		Data: comment_notify.Data{
			Title:   post.Title,
			Link:    fmt.Sprintf(`%s#comment-%d`, post.Link, c.ID),
			Date:    time.Unix(int64(c.Date), 0).In(loc).Format(time.RFC3339),
			Author:  c.Author,
			Content: c.Source,
		},
		Email:    c.Email,
		HomePage: c.URL,
	}
	t.s.cmtntf.NotifyAdmin(&data)

	// 通知文章作者：访客或者非自己
	if c.UserID == 0 || c.UserID != post.UserId {
		r := comment_notify.Recipient{}
		r.BarkToken = pu.BarkToken
		r.Email.Name = pu.Nickname
		r.Email.Address = pu.Email
		t.s.cmtntf.NotifyPostAuthor(&data.Data, r)
	}

	// 通知其它上级回复：所有上级中排除特定人后……
	var parents []models.Comment
	for parentID := c.Parent; parentID > 0; {
		var parent models.Comment
		t.s.tdb.From(parent).
			Select("id,author,email,parent,user_id").
			Where("id=?", parentID).
			MustFind(&parent)
		parents = append(parents, parent)
		parentID = parent.Parent
	}
	if len(parents) == 0 {
		return nil
	}

	// 有几类特定人需要排除：
	// 1. 管理员（已经发过即时通知）
	// 2. 文章作者，前面已经通知过。
	// 3. 没有邮箱的
	// 4. 评论者自己
	//    1. 如果没有登录，判断邮箱相同
	//    2. 如果已经登录，判断用户ID相同
	distinct := map[string]bool{}
	recipients := []comment_notify.Recipient{}
	for _, parent := range parents {
		if parent.UserID == int32(auth.AdminID) {
			continue
		}
		if parent.UserID == post.UserId {
			continue
		}
		if parent.Email == `` {
			continue
		}
		if parent.UserID == 0 {
			if parent.Email == c.Email {
				continue
			}
		} else {
			if parent.UserID == c.UserID {
				continue
			}
		}

		r := comment_notify.Recipient{}

		// 首先从评论本身拿邮件，
		email := parent.Email
		// 而如果是登录用户，则优先使用配置的邮箱/即时通知。
		if parent.UserID > 0 {
			parentUser, err := t.s.auth.GetUserByID(context.TODO(), int64(parent.UserID))
			if err == nil {
				email = parentUser.Email
				r.BarkToken = parentUser.BarkToken
			}
		}
		r.Email.Name = parent.Author // or parentUser.Nickname
		r.Email.Address = email

		once := strings.ToLower(cmp.Or(email, r.BarkToken))
		if _, ok := distinct[once]; !ok {
			distinct[once] = true
			recipients = append(recipients, r)
		}
	}

	if len(recipients) == 0 {
		return nil
	}

	t.s.cmtntf.NotifyGuests(&data.Data, recipients)

	return nil
}

func (s *Service) deletePostComments(_ context.Context, postID int64) {
	s.tdb.From(models.Comment{}).Where(`post_id=?`, postID).MustDelete()
}

// 请保持文章和评论的代码同步。
// NOTE：评论不需要检测权限，UpdateComment 会检测。
func (s *Service) CheckCommentTaskListItems(ctx context.Context, in *proto.CheckTaskListItemsRequest) (*proto.CheckTaskListItemsResponse, error) {
	p := s.getComment2(int64(in.Id)).ToProto(s.setCommentExtraFields(ctx, &proto.PostContentOptions{}))
	updated, err := s.applyTaskChecks(p.Modified, p.SourceType, p.Source, in)
	if err != nil {
		return nil, err
	}

	p.Source = string(updated)

	updateRequest := proto.UpdateCommentRequest{
		Comment: p,
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{
				`source_type`,
				`source`,
				`modified`,
			},
		},
	}
	updatedComment, err := s.UpdateComment(ctx, &updateRequest)
	if err != nil {
		return nil, err
	}

	return &proto.CheckTaskListItemsResponse{
		ModificationTime: updatedComment.Modified,
	}, nil
}

func (s *Service) ResolveAvatar(id uint32) (string, time.Time, io.ReadSeekCloser) {
	userID, email := s.avatarCache.Resolve(id)
	if email != `` {
		return email, time.Time{}, nil
	}
	user, err := s.auth.GetUserByID(context.Background(), int64(userID))
	if err != nil || len(user.Avatar.Data) <= 0 {
		return ``, time.Time{}, nil
	}
	return ``, time.Unix(user.UpdatedAt, 0), _ByteReadCloser{
		Reader: bytes.NewReader(user.Avatar.Data),
	}
}

type _ByteReadCloser struct {
	*bytes.Reader
}

func (_ByteReadCloser) Close() error { return nil }
