package e2e_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
)

func TestGetNonExistentPage(t *testing.T) {
	r := Serve(context.Background())
	expectHTTPGetWithStatusCode(r, `/page-that-does-not-exist`, 404)
}

func TestGetNonExistentPost(t *testing.T) {
	r := Serve(context.Background())
	expectHTTPGetWithStatusCode(r, `/2147483647/`, 404)
}

func TestNoAccessToPost(t *testing.T) {
	r := Serve(context.Background())
	p := utils.Must1(r.client.Blog.CreatePost(r.admin, &proto.Post{
		SourceType: `markdown`,
		Source:     `# 测试私密文章。`,
		Status:     `draft`,
	}))
	expectHTTPGetWithStatusCode(r, fmt.Sprintf(`/%d/`, p.Id), 404)
}

func TestModTime(t *testing.T) {
	r := Serve(context.Background())
	expect304(r, `/style.css`)
	expect304(r, `/admin/script.js`)
}
