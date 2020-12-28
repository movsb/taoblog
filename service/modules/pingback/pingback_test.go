package pingback

import (
	"context"
	"fmt"
	"testing"
)

func TestFindServer(t *testing.T) {
	t.SkipNow()
	s, e := findServer(context.TODO(), `https://coolshell.cn/articles/21263.html`)
	fmt.Println(s, e)
}

func TestPing(t *testing.T) {
	t.SkipNow()
	Ping(context.TODO(), `https://www.google.com`, `https://coolshell.cn/articles/21263.html`)
}
