package service

import (
	"context"
	"log"
	"time"

	"github.com/blugelabs/bluge/search/highlight"
	"github.com/movsb/taoblog/protocols"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taoblog/service/modules/search"
)

// TODO 权限
func (s *Service) SearchPosts(ctx context.Context, in *protocols.SearchPostsRequest) (*protocols.SearchPostsResponse, error) {
	searcher := s.searcher.Load()
	if searcher == nil {
		return &protocols.SearchPostsResponse{
			Initialized: false,
		}, nil
	}
	posts, err := searcher.SearchPosts(ctx, in.Search)
	if err != nil {
		return nil, err
	}
	highlighter := highlight.NewHTMLHighlighterTags(`<b class="highlight">`, `</b>`)
	rspPosts := []*protocols.SearchPostsResponse_Post{}
	for _, post := range posts {
		rspPosts = append(rspPosts, &protocols.SearchPostsResponse_Post{
			Id:      int32(post.ID),
			Title:   highlighter.BestFragment(post.Locations[`title`], []byte(post.Title)),
			Content: highlighter.BestFragment(post.Locations[`content`], []byte(post.Content)),
		})
	}
	return &protocols.SearchPostsResponse{
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
	var posts models.Posts
	err := s.tdb.Model(posts).Where(`modified > ?`, *lastCheck).Find(&posts)
	if err != nil {
		log.Println(err)
		return
	}
	if len(posts) <= 0 {
		return
	}
	var posts2 []search.Post
	for _, p := range posts {
		posts2 = append(posts2, search.Post{
			ID:      p.ID,
			Title:   p.Title,
			Content: p.Source,
		})
	}
	err = engine.IndexPosts(ctx, posts2)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("indexed %d posts\n", len(posts2))
	*lastCheck = now.Unix()
}
