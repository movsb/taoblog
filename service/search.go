package service

import (
	"context"
	"log"
	"time"

	"github.com/blugelabs/bluge/search/highlight"
	"github.com/movsb/taoblog/modules/auth"
	co "github.com/movsb/taoblog/protocols/go/handy/content_options"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/modules/search"
)

func (s *Service) SearchPosts(ctx context.Context, in *proto.SearchPostsRequest) (*proto.SearchPostsResponse, error) {
	searcher := s.searcher.Load()
	if searcher == nil {
		// 首次搜索时初始化搜索引擎。
		s.onceInitSearcher.Do(func() {
			ch := make(chan struct{})
			go s.runSearchEngine(s.ctx, ch)
			for i := range 30 {
				select {
				case <-time.After(time.Second):
					log.Println(`等待搜索引擎初始化完成：`, i+1)
				case <-ch:
					return
				}
			}
		})
		searcher = s.searcher.Load()
		if searcher == nil {
			return &proto.SearchPostsResponse{
				Initialized: false,
			}, nil
		}
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

func (s *Service) runSearchEngine(ctx context.Context, ch chan<- struct{}) {
	engine, err := search.NewEngine(&s.cfg.Search)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		engine.Close()
		s.searcher.Store(nil)
	}()

	var lastCheck int64
	s.reIndex(ctx, engine, &lastCheck)

	s.searcher.Store(engine)
	ch <- struct{}{}

	const scanInterval = time.Minute * 1

	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	for loop := true; loop; {
		select {
		case <-ticker.C:
			s.reIndex(ctx, engine, &lastCheck)
		case <-ctx.Done():
			loop = false
		}
	}
}

func (s *Service) reIndex(ctx context.Context, engine *search.Engine, lastCheck *int64) {
	now := time.Now()
	rsp, err := s.ListPosts(auth.SystemForLocal(ctx), &proto.ListPostsRequest{
		GetPostOptions: &proto.GetPostOptions{
			ContentOptions: co.For(co.SearchIndex),
			WithLink:       proto.LinkKind_LinkKindRooted,
		},
		ModifiedNotBefore: int32(*lastCheck),
	})
	if err != nil {
		log.Println(err)
		return
	}
	posts := rsp.Posts
	if len(posts) <= 0 {
		return
	}
	err = engine.IndexPosts(ctx, posts)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("indexed %d posts\n", len(posts))
	*lastCheck = now.Unix()
}
