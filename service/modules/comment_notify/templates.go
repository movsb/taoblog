package comment_notify

const adminTemplate = `
<b>您的博文“{{.Title}}”有新的评论啦！</b><br/><br/>

<b>链接：</b><a href="{{.Link}}">{{.Link}}</a><br/>
<b>作者：</b>{{.Author}}<br/>
<b>邮箱：</b>{{.Email}}<br/>
<b>网址：</b>{{.HomePage}}<br/>
<b>时间：</b>{{.Date}}<br/>
<b>内容：</b>{{.Content}}<br/>
`

const guestTemplate = `
<b>您在博文“{{.Title}}”的评论有新的回复啦！</b><br/><br/>

<b>链接：</b><a href="{{.Link}}">{{.Link}}</a><br/>
<b>作者：</b>{{.Author}}<br/>
<b>时间：</b>{{.Date}}<br/>
<b>内容：</b>{{.Content}}<br/>

<br/>该邮件为系统自动发出，请勿直接回复该邮件。<br/>
`
