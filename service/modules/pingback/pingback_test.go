package pingback

import (
	"context"
	"fmt"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func init() {
	utils.InitTestLogger()
}

func TestFindServer(t *testing.T) {
	t.SkipNow()
	s, e := findServer(context.TODO(), `https://coolshell.cn/articles/21263.html`)
	fmt.Println(s, e)
}

func TestPing(t *testing.T) {
	t.SkipNow()
	Ping(context.TODO(), `http://localhost:2564/849/`, `http://localhost:2564/849/`)
}
