package main

import (
	"bytes"
	"fmt"
	"html"
	"log"
	"strings"

	"github.com/movsb/taoblog/server/modules/mailer"
)

func doSendMailAsync(recipient, nickname, subject, body string) {
	log.Println("send_mail:", nickname, recipient, subject, body)

	cfg := strings.SplitN(config.mail, "/", 3)
	if len(cfg) != 3 {
		panic("bad mail config")
	}

	go func() {
		mc, err := mailer.DialTLS(cfg[0])
		if err != nil {
			log.Println(err)
			return
		}
		defer mc.Quit()
		if err = mc.Auth(cfg[1], cfg[2]); err != nil {
			log.Println(err)
			return
		}
		if err = mc.SetFrom("博客评论", cfg[1]); err != nil {
			log.Println("SetFrom:", err)
			return
		}
		if err = mc.AddTo(nickname, recipient); err != nil {
			log.Println("AddTo:", recipient, err)
			return
		}
		if err = mc.Send(subject, body); err != nil {
			log.Println(err)
			return
		}
	}()
}

func doNotifyAdmin(tx Querier, cmt *Comment, postTitle string) {
	body := bytes.NewBuffer(nil)

	write := func(f string, args ...interface{}) {
		body.WriteString(fmt.Sprintf(f, args...))
	}

	write(`<b>您的博文“%s”有新的评论啦！</b><br/><br/>`, html.EscapeString(postTitle))

	link := "https://" + optmgr.GetDef(tx, "home", "localhost") + "/?p=" + fmt.Sprint(cmt.PostID) + "#comments"
	write(`<b>链接：</b>%s<br/>`, link)
	write(`<b>作者：</b>%s<br/>`, html.EscapeString(string(cmt.Author)))
	write(`<b>邮箱：</b>%s<br/>`, cmt.EMail)
	write(`<b>网址：</b>%s<br/>`, html.EscapeString(cmt.URL))
	write(`<b>时间：</b>%s<br/>`, cmt.Date)
	write(`<b>内容：</b>%s<br/>`, html.EscapeString(cmt.Content))

	doSendMailAsync(
		optmgr.GetDef(tx, "email", ""),
		optmgr.GetDef(tx, "nickname", ""),
		fmt.Sprintf("[新博文评论] %s", postTitle),
		body.String(),
	)
}

type ParentInfo struct {
	ID     int64
	Author string
	Email  string
}

func doNotifyUser(tx Querier, cmt *Comment, postTitle string, parent ParentInfo) {
	link := "https://" + optmgr.GetDef(tx, "home", "localhost") + "/?p=" + fmt.Sprint(cmt.PostID) + "#comments"

	body := bytes.NewBuffer(nil)

	write := func(f string, args ...interface{}) {
		body.WriteString(fmt.Sprintf(f, args...))
	}

	write(`<b>您在博文“%s”的评论有新的回复啦！</b><br/><br/>`, postTitle)
	write(`<b>链接：</b>%s<br/>`, link)
	write(`<b>作者：</b>%s<br/>`, html.EscapeString(string(cmt.Author)))
	write(`<b>时间：</b>%s<br/>`, cmt.Date)
	write(`<b>内容：</b>%s<br>`, html.EscapeString(cmt.Content))
	write(`<br/>该邮件为系统自动发出，请勿直接回复该邮件。<br/>`)

	subject := "[回复评论] " + postTitle

	doSendMailAsync(parent.Email, parent.Author, subject, body.String())
}

func doNotify(tx Querier, cmt *Comment) {
	var err error

	adminEmail := optmgr.GetDef(tx, "email", "")

	var postTitle string
	if err = postmgr.GetVars(tx, "title", "id="+fmt.Sprint(cmt.PostID), &postTitle); err != nil {
		log.Println(err)
		return
	}

	if cmt.EMail != adminEmail {
		doNotifyAdmin(tx, cmt, postTitle)
	}

	parents := []ParentInfo{}

	parentID := cmt.Parent
	for parentID > 0 {
		var parentInfo ParentInfo
		if err = cmtmgr.GetVars(tx, "id,author,email,parent", "id="+fmt.Sprint(parentID),
			&parentInfo.ID, &parentInfo.Author, &parentInfo.Email, &parentID); err != nil {
			log.Println(err)
			return
		}
		parents = append(parents, parentInfo)
		break // No notify upper user currently
	}

	for _, parent := range parents {
		doNotifyUser(tx, cmt, postTitle, parent)
	}
}
