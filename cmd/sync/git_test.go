package sync

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/movsb/taoblog/cmd/client"
	"github.com/movsb/taoblog/protocols/clients"
)

func TestAll(t *testing.T) {
	t.SkipNow()
	config := client.HostConfig{
		Home:  `https://blog.home.twofei.com`,
		Token: `12345678`,
	}
	client := clients.NewProtoClient(config.Home, config.Token)
	g := New(client, Credential{}, "/tmp/", false)
	posts, err := g.getUpdatedPosts(time.Now(), time.Now())
	log.Println(posts, err)
}

func TestFind(t *testing.T) {
	t.SkipNow()
	t.Log(findPostByID(os.DirFS(`/Users/tao/Documents/posts`), 900))
}
