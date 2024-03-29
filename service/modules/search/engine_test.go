package search

import (
	"context"
	"fmt"
	"testing"

	"github.com/blugelabs/bluge/search/highlight"
)

func TestEngine(t *testing.T) {
	cfg := DefaultConfig()
	engine, err := NewEngine(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.IndexPosts(context.TODO(), []Post{
		{ID: 1, Title: `标题`, Content: `桃子爱雪舞桃子不爱雪舞`},
	}); err != nil {
		t.Fatal(err)
	}
	result, err := engine.SearchPosts(context.TODO(), `标桃雪`)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)

	highlighter := highlight.NewHTMLHighlighter()
	for _, post := range result {
		s := highlighter.BestFragment(post.Locations[`title`], []byte(post.Title))
		fmt.Println(s)
		s = highlighter.BestFragment(post.Locations[`content`], []byte(post.Content))
		fmt.Println(s)
	}
}
