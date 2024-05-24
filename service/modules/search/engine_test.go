package search

import (
	"context"
	"fmt"
	"testing"

	"github.com/blugelabs/bluge/search/highlight"
	proto "github.com/movsb/taoblog/protocols"
	search_config "github.com/movsb/taoblog/service/modules/search/config"
)

func TestEngine(t *testing.T) {
	cfg := search_config.DefaultConfig()
	engine, err := NewEngine(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.IndexPosts(context.TODO(), []*proto.Post{
		{Id: 1, Title: `标题`, Source: `昔我往矣，杨柳依依。今我来思，雨雪霏霏。`},
	}); err != nil {
		t.Fatal(err)
	}
	result, err := engine.SearchPosts(context.TODO(), `杨柳依依`)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)

	highlighter := highlight.NewHTMLHighlighter()
	for _, post := range result {
		s := highlighter.BestFragment(post.Locations[`title`], []byte(post.Post.Title))
		fmt.Println(s)
		s = highlighter.BestFragment(post.Locations[`content`], []byte(post.Post.Source))
		fmt.Println(s)
	}
}
