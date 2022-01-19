package search

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/blugelabs/bluge"
	"github.com/blugelabs/bluge/search"
)

type Engine struct {
	cfg    *Config
	writer *bluge.Writer
	closed bool
	mu     sync.RWMutex
}

func NewEngine(cfg *Config) (*Engine, error) {
	engine := &Engine{
		cfg: cfg,
	}
	return engine, nil
}

// TODO: 私有文章不允许非登录用户搜索
type Post struct {
	ID        int64
	Title     string
	Content   string
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

func (e *Engine) IndexPosts(ctx context.Context, posts []Post) (err error) {
	writer, err := e.getWriter()
	if err != nil {
		return err
	}
	batch := bluge.NewBatch()
	for _, post := range posts {
		id := strconv.Itoa(int(post.ID))
		doc := bluge.NewDocument(id)
		{
			titleField := bluge.NewTextField(`title`, post.Title)
			titleField.FieldOptions = bluge.Index | bluge.Store | bluge.HighlightMatches
			doc.AddField(titleField)
		}
		{
			contentField := bluge.NewTextField(`content`, post.Content)
			contentField.FieldOptions = bluge.Index | bluge.Store | bluge.HighlightMatches
			doc.AddField(contentField)
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

func (e *Engine) SearchPosts(ctx context.Context, search string) (posts []Post, err error) {
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
	matchContent.SetField(`content`)
	query.AddShould(matchContent)

	searchRequest := bluge.NewTopNSearch(10, query).IncludeLocations()
	dmi, err := reader.Search(ctx, searchRequest)
	if err != nil {
		return nil, err
	}
	var result []Post
	next, err := dmi.Next()
	for err == nil && next != nil {
		var post Post
		err = next.VisitStoredFields(func(field string, value []byte) bool {
			switch field {
			case `_id`:
				id, _ := strconv.Atoi(string(value))
				post.ID = int64(id)
			case `title`:
				post.Title = string(value)
			case `content`:
				post.Content = string(value)
			}
			return true
		})
		if err == nil {
			post.Locations = next.Locations
			result = append(result, post)
			next, err = dmi.Next()
		}
	}
	if err != nil {
		return nil, err
	}

	log.Println("结果数", len(result))
	return result, nil
}
