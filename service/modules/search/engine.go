package search

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
	"github.com/movsb/taoblog/modules/auth"
	"github.com/movsb/taoblog/protocols/go/proto"
	search_config "github.com/movsb/taoblog/service/modules/search/config"
)

type Engine struct {
	cfg    *search_config.Config
	writer *bluge.Writer
	closed bool
	mu     sync.RWMutex
}

func NewEngine(cfg *search_config.Config) (*Engine, error) {
	engine := &Engine{
		cfg: cfg,
	}
	return engine, nil
}

type SearchResult struct {
	Post      proto.Post
	Locations search.FieldTermLocationMap
}

func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	writer, err := e.getWriter()
	if err != nil {
		return err
	}
	writer.Close()
	e.closed = true
	return nil
}

func (e *Engine) IndexPosts(ctx context.Context, posts []*proto.Post) (err error) {
	writer, err := e.getWriter()
	if err != nil {
		return err
	}
	batch := bluge.NewBatch()
	for _, post := range posts {
		id := fmt.Sprint(post.Id)
		doc := bluge.NewDocument(id)
		{
			titleField := bluge.NewTextField(`title`, post.Title)
			titleField.FieldOptions = bluge.Index | bluge.Store | bluge.HighlightMatches
			doc.AddField(titleField)
		}
		{
			contentField := bluge.NewTextField(`source`, post.Source)
			contentField.FieldOptions = bluge.Index | bluge.Store | bluge.HighlightMatches
			doc.AddField(contentField)
		}
		{
			statusField := bluge.NewTextField(`status`, post.Status)
			statusField.FieldOptions = bluge.Index | bluge.Store
			doc.AddField(statusField)
		}
		batch.Update(doc.ID(), doc)
	}
	if err = writer.Batch(batch); err != nil {
		return err
	}
	return nil
}

func (e *Engine) getWriter() (*bluge.Writer, error) {
	e.mu.RLock()
	if e.writer != nil {
		e.mu.RUnlock()
		return e.writer, nil
	}
	e.mu.RUnlock()

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.closed {
		return nil, fmt.Errorf(`closed`)
	}

	if e.writer != nil {
		return e.writer, nil
	}

	var writerConfig bluge.Config
	if e.cfg.InMemory {
		writerConfig = bluge.InMemoryOnlyConfig()
	} else {
		writerConfig = bluge.DefaultConfig(e.cfg.Paths.Data)
	}
	writer, err := bluge.OpenWriter(writerConfig)
	if err != nil {
		return nil, err
	}

	e.writer = writer
	return writer, nil
}

func (e *Engine) SearchPosts(ctx context.Context, search string) (posts []*SearchResult, err error) {
	writer, err := e.getWriter()
	if err != nil {
		return nil, err
	}
	reader, err := writer.Reader()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	query := bluge.NewBooleanQuery()
	query.SetMinShould(1)

	matchTitle := bluge.NewMatchQuery(search)
	matchTitle.SetField(`title`)
	query.AddShould(matchTitle)

	matchContent := bluge.NewMatchQuery(search)
	matchContent.SetField(`source`)
	query.AddShould(matchContent)

	if user := auth.Context(ctx).User; !user.IsAdmin() && !user.IsSystem() {
		matchStatus := bluge.NewTermQuery(`public`)
		matchStatus.SetField(`status`)
		query.AddMust(matchStatus)
	}

	searchRequest := bluge.NewTopNSearch(10, query).IncludeLocations()
	dmi, err := reader.Search(ctx, searchRequest)
	if err != nil {
		return nil, err
	}
	var result []*SearchResult
	next, err := dmi.Next()
	for err == nil && next != nil {
		var post SearchResult
		err = next.VisitStoredFields(func(field string, value []byte) bool {
			switch field {
			case `_id`:
				id, _ := strconv.Atoi(string(value))
				post.Post.Id = int64(id)
			case `title`:
				post.Post.Title = string(value)
			case `source`:
				post.Post.Source = string(value)
			case `status`:
				post.Post.Status = string(value)
			}
			return true
		})
		if err == nil {
			post.Locations = next.Locations
			result = append(result, &post)
			next, err = dmi.Next()
		}
	}
	if err != nil {
		return nil, err
	}

	log.Println("结果数", len(result))
	return result, nil
}
