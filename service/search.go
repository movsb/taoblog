package service

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/blugelabs/bluge/search/highlight"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/micros/auth/user"
	"github.com/movsb/taoblog/service/modules/search/inverted"
)

func (s *Service) SearchPosts(ctx context.Context, in *proto.SearchPostsRequest) (*proto.SearchPostsResponse, error) {
	searcher, err := s.waitForSearchEngine(ctx)
	if err != nil {
		return nil, err
	}
	if searcher == nil {
		return &proto.SearchPostsResponse{Initialized: false}, nil
	}

	posts, err := searcher.SearchPosts(ctx, in.Search)
	if err != nil {
		return nil, err
	}
	highlighter := highlight.NewHTMLHighlighterTags(`<b class="highlight">`, `</b>`)
	rspPosts := []*proto.SearchPostsResponse_Post{}
	for _, post := range posts {
		rspPosts = append(rspPosts, &proto.SearchPostsResponse_Post{
			Id:      int32(post.Post.Id),
			Title:   highlighter.BestFragment(post.Locations[`title`], []byte(post.Post.Title)),
			Content: highlighter.BestFragment(post.Locations[`source`], []byte(post.Post.Source)),
		})
	}
	return &proto.SearchPostsResponse{
		Posts:       rspPosts,
		Initialized: true,
	}, nil
}

func (s *Service) waitForSearchEngine(ctx context.Context) (*inverted.Engine, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := s.ctx.Err(); err != nil {
		return nil, err
	}
	if searcher := s.searcher.Load(); searcher != nil {
		return searcher, nil
	}

	// sync.Once 只负责启动，所有请求各自等待就绪、超时或取消。
	s.onceInitSearcher.Do(func() {
		s.searcherReady = make(chan struct{})
		go s.runSearchEngine(s.ctx, s.searcherReady)
	})
	timer := time.NewTimer(30 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.ctx.Done():
		return nil, s.ctx.Err()
	case <-timer.C:
		return nil, nil
	case <-s.searcherReady:
		if err := s.ctx.Err(); err != nil {
			return nil, err
		}
		return s.searcher.Load(), s.searcherInitErr
	}
}

func (s *Service) runSearchEngine(ctx context.Context, ch chan<- struct{}) {
	// 关闭 channel 广播结果，调用方已超时或取消也不会阻塞后台任务。
	notifyReady := sync.OnceFunc(func() { close(ch) })
	defer notifyReady()
	engine, err := inverted.NewEngine(&s.cfg.Search)
	if err != nil {
		s.searcherInitErr = fmt.Errorf("搜索引擎初始化失败：%w", err)
		log.Println(s.searcherInitErr)
		return
	}
	defer func() {
		s.searcher.Store(nil)
		if err := engine.Close(); err != nil {
			log.Println(`关闭搜索引擎失败：`, err)
		}
	}()

	var lastCheck int64
	if err := s.reIndex(ctx, engine, &lastCheck); err != nil {
		s.searcherInitErr = fmt.Errorf("搜索引擎初次索引失败：%w", err)
		log.Println(s.searcherInitErr)
		return
	}
	if err := ctx.Err(); err != nil {
		s.searcherInitErr = err
		return
	}

	s.searcher.Store(engine)
	notifyReady()

	const scanInterval = time.Minute * 1

	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for loop := true; loop; {
		select {
		case <-ticker.C:
			if err := s.reIndex(ctx, engine, &lastCheck); err != nil {
				log.Println(`更新搜索索引失败：`, err)
			}
		case <-ctx.Done():
			loop = false
		}
	}
}

func (s *Service) reIndex(ctx context.Context, engine *inverted.Engine, lastCheck *int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	now := time.Now()
	rsp, err := s.ListPosts(user.SystemForLocal(ctx), &proto.ListPostsRequest{
		GetPostOptions: &proto.GetPostOptions{
			ContentOptions: co.For(co.SearchIndex),
			WithLink:       proto.LinkKind_LinkKindRooted,
		},
		ModifiedNotBefore: int32(*lastCheck),
	})
	if err != nil {
		return err
	}
	posts := rsp.Posts
	if len(posts) <= 0 {
		return nil
	}
	err = engine.IndexPosts(ctx, posts)
	if err != nil {
		return err
	}
	log.Printf("indexed %d posts\n", len(posts))
	*lastCheck = now.Unix()
	return nil
}
