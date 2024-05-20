package sync

import (
	"log"
	"os"
	"testing"

	"github.com/movsb/taoblog/cmd/client"
)

func TestAll(t *testing.T) {
	t.SkipNow()
	g := New(client.HostConfig{
		API:   `https://blog.home.twofei.com/v3`,
		Token: `12345678`,
	}, "/tmp/")
	posts, err := g.getUpdatedPosts()
	log.Println(posts, err)
}

func TestFind(t *testing.T) {
	t.SkipNow()
	t.Log(findPostByID(os.DirFS(`/Users/tao/Documents/posts`), 900))
}
