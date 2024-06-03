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
		return &proto.SearchPostsResponse{
			Initialized: false,
		}, nil
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

func (s *Service) RunSearchEngine(ctx context.Context) {
	time.Sleep(s.cfg.Search.InitialDelay)

	engine, err := search.NewEngine(&s.cfg.Search)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		engine.Close()
		s.searcher.Store(nil)
	}()

	s.searcher.Store(engine)

	var lastCheck int64
	s.reIndex(ctx, engine, &lastCheck)

	ticker := time.NewTicker(s.cfg.Search.ScanInterval)
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
	rsp, err := s.ListPosts(auth.SystemAdmin(ctx), &proto.ListPostsRequest{
		ContentOptions:    co.For(co.SearchIndex),
		WithLink:          proto.LinkKind_LinkKindRooted,
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
