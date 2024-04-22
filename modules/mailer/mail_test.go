package mailer

import (
	"testing"
)

func Test(t *testing.T) {
	t.SkipNow()

	m, err := DialTLS("smtp.qq.com:465")
	if err != nil {
		t.Fatal(err)
	}

	defer m.Quit()

	if err = m.Auth("blog@twofei.com", "***"); err != nil {
		t.Fatal(err)
	}

	if err = m.SetFrom("博客评论", "blog@twofei.com"); err != nil {
		t.Fatal(err)
	}

	if err = m.AddTo("自己", "anhbk@qq.com"); err != nil {
		t.Fatal(err)
	}

	if err = m.Send("主题", "内容"); err != nil {
		t.Fatal(err)
	}

	m.Quit()
}
