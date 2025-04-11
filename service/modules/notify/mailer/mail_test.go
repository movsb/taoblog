package mailer_test

import (
	"testing"

	"github.com/movsb/taoblog/service/modules/notify/mailer"
)

func Test(t *testing.T) {
	t.SkipNow()

	m := mailer.NewMailer(`smtp.qq.com:587`, `blog@twofei.com`, `***`)
	if err := m.Send(`主题`, `标题`, `博客评论`, []mailer.User{{`自己`, `anhbk@qq.com`}}); err != nil {
		t.Fatal(err)
	}
}
