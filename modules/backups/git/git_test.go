package backups_git

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/movsb/taoblog/protocols/clients"
)

func TestAll(t *testing.T) {
	t.SkipNow()
	home := `https://blog.home.twofei.com`
	token := `12345678`
	client := clients.NewProtoClientFromHome(home, token)
	g := New(context.Background(), client, false)
	posts, err := g.getUpdatedPosts(time.Now(), time.Now())
	log.Println(posts, err)
}

func TestFind(t *testing.T) {
	t.SkipNow()
	t.Log(findPostByID(os.DirFS(`/Users/tao/Documents/posts`), 900, nil))
}

func TestClone(t *testing.T) {
	t.SkipNow()
	dir, repo, err := clone(
		context.Background(),
		`https://github.com/rsc/gitfs`,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = repo
	t.Logf(`克隆到：%s`, dir)
}
