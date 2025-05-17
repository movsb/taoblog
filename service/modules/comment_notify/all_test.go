package comment_notify

import (
	"os"
	"testing"

	"github.com/movsb/taoblog/modules/utils"
)

func TestTemplates(t *testing.T) {
	d := AdminData{
		Data: Data{
			Title:   `文章标题`,
			Link:    `https://example.com/1/`,
			Date:    `2025-05-14`,
			Author:  `评论者昵称`,
			Content: `评论内容`,
		},
		Email:    `someone@example.com`,
		HomePage: `https://example.com`,
	}
	t.Log(`给文章作者的邮件通知：`)
	utils.Must(authorMailTmpl.Execute(os.Stdout, d.Data))
	t.Log(`能评论者的邮件通知：`)
	utils.Must(guestMailTmpl.Execute(os.Stdout, d.Data))
	t.Log(`给站长或者登录者的即时通知：`)
	utils.Must(barkTmpl.Execute(os.Stdout, d))
}
