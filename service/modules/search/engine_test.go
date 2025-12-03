package search

import (
	"context"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"github.com/blugelabs/bluge/search/highlight"
	"github.com/movsb/taoblog/modules/auth/user"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
	"github.com/movsb/taoblog/service/models"
	search_config "github.com/movsb/taoblog/service/modules/search/config"
)

func TestEngine(t *testing.T) {
	cfg := search_config.DefaultConfig()
	engine, err := NewEngine(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.IndexPosts(context.TODO(), []*proto.Post{
		{Id: 1, Title: `标题`, Source: `内容测试`, Status: models.PostStatusPublic, UserId: 1},
		{Id: 2, Title: `标题`, Source: `容内测试`, Status: models.PostStatusPublic, UserId: 1},
	}); err != nil {
		t.Fatal(err)
	}
	result, err := engine.SearchPosts(user.GuestForLocal(context.TODO()), `内容`)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(result)

	highlighter := highlight.NewHTMLHighlighterTags(`<b class="highlight">`, `</b>`)
	for _, post := range result {
		s := highlighter.BestFragment(post.Locations[`title`], []byte(post.Post.Title))
		fmt.Println(s)
		s = highlighter.BestFragment(post.Locations[`source`], []byte(post.Post.Source))
		fmt.Println(s)
	}
}

func TestPerm(t *testing.T) {
	cfg := search_config.DefaultConfig()
	engine, err := NewEngine(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	u1 := &user.User{User: &models.User{ID: 1}}
	u2 := &user.User{User: &models.User{ID: 2}}
	u3 := &user.User{User: &models.User{ID: 3}}
	if err := engine.IndexPosts(context.TODO(), []*proto.Post{
		{Id: 1, Title: `公开文章`, Status: models.PostStatusPublic, UserId: int32(u1.ID)},
		{Id: 2, Title: `私有文章`, Status: models.PostStatusPrivate, UserId: int32(u2.ID)},
		{Id: 3, Title: `草稿文章`, Status: models.PostStatusDraft, UserId: int32(u3.ID)},
	}); err != nil {
		t.Fatal(err)
	}

	testCases := []struct {
		user   context.Context
		search string
		ids    []int
	}{
		{user.GuestForLocal(context.TODO()), `文章`, []int{1}},
		{user.TestingUserContextForServer(u1), `文章`, []int{1}},
		{user.TestingUserContextForServer(u2), `文章`, []int{1, 2}},
		{user.TestingUserContextForServer(u3), `文章`, []int{1, 3}},
	}

	for _, tc := range testCases {
		result, err := engine.SearchPosts(tc.user, tc.search)
		if err != nil {
			t.Fatal(err)
		}
		ids := utils.Map(result, func(p *SearchResult) int {
			return int(p.Post.Id)
		})
		slices.Sort(ids)
		if !reflect.DeepEqual(ids, tc.ids) {
			user := user.Context(tc.user).User
			t.Errorf("user: %d, search: %s, expect ids: %v, got: %v", user.ID, tc.search, tc.ids, ids)
			continue
		}
	}
}
