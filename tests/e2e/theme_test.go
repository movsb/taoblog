package e2e_test

import (
	"fmt"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/protocols/go/proto"
)

func TestGetNonExistentPage(t *testing.T) {
	expectHTTPGetWithStatusCode(`/page-that-does-not-exist`, 404)
}

func TestGetNonExistentPost(t *testing.T) {
	expectHTTPGetWithStatusCode(`/2147483647/`, 404)
}

func TestNoAccessToPost(t *testing.T) {
	p := utils.Must1(client.Blog.CreatePost(admin, &proto.Post{
		SourceType: `markdown`,
		Source:     `# 测试私密文章。`,
		Status:     `draft`,
	}))
	expectHTTPGetWithStatusCode(fmt.Sprintf(`/%d/`, p.Id), 404)
}
