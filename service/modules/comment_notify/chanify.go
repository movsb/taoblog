package comment_notify

import (
	"bytes"
	"log"
	"text/template"

	"github.com/movsb/pkg/notify"
)

var _chanifyTemplate = template.Must(template.New(`chanify`).Parse(`
{{- .Content }}

文章链接：{{ .Link }}
评论作者：{{ .Author }}
评论日期：{{ .Date }}
作者邮箱：{{ .Email }}
作者主页：{{ .HomePage }}
`))

func executeChanifyTemplate(data *AdminData) string {
	b := bytes.NewBuffer(nil)
	if err := _chanifyTemplate.Execute(b, data); err != nil {
		log.Println(err)
		return ""
	}
	return b.String()
}

// Chanify ...
func (cn *CommentNotifier) Chanify(data *AdminData) {
	token := cn.Config.Push.Chanify.Token
	if token == "" {
		return
	}
	ch := notify.NewOfficialChanify(token)
	ch.Send(data.Title, executeChanifyTemplate(data), true)
}

func (cn *CommentNotifier) ChanifyPlain(title, content string) {
	token := cn.Config.Push.Chanify.Token
	if token == "" {
		return
	}
	ch := notify.NewOfficialChanify(token)
	ch.Send(title, content, true)
}
