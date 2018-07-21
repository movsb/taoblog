package main

import (
	"bytes"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"./internal/file_managers"
	"./internal/mailer"
	"./internal/utils/datetime"
	"./internal/utils/hooks"
)

type xConfig struct {
	base     string
	listen   string
	username string
	password string
	database string
	key      string
	files    string
	fileHost string
	mail     string
}

var gkey string
var config xConfig
var gdb *sql.DB
var tagmgr *xTagManager
var postmgr *xPostManager
var optmgr *OptionManager
var auther *GenericAuth
var uploadmgr *FileUpload
var backupmgr *BlogBackup
var cmtmgr *CommentManager
var postcmtsmgr *PostCommentsManager
var fileredir *FileRedirect

var gHooks *hooks.HookManager

func auth(c *gin.Context, finish bool) bool {
	if auther.AuthHeader(c) || auther.AuthCookie(c) {
		return true
	}
	if finish {
		EndReq(c, false, "auth error")
	}
	return false
}

func main() {
	flag.StringVar(&config.listen, "listen", "127.0.0.1:2564", "the port to which the server listen")
	flag.StringVar(&config.username, "username", "taoblog", "the database username")
	flag.StringVar(&config.password, "password", "taoblog", "the database password")
	flag.StringVar(&config.database, "database", "taoblog", "the database name")
	flag.StringVar(&config.key, "key", "", "api key")
	flag.StringVar(&config.base, "base", ".", "taoblog directory")
	flag.StringVar(&config.files, "files", ".", "the files folder")
	flag.StringVar(&config.fileHost, "file-host", "//localhost", "the backup file host")
	flag.StringVar(&config.mail, "mail", "//", "example.com:465/user@example.com/password")
	flag.Parse()

	if config.key == "" {
		panic("invalid key")
	}

	var err error
	dataSource := fmt.Sprintf("%s:%s@/%s", config.username, config.password, config.database)
	gdb, err = sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}

	gdb.SetMaxIdleConns(10)

	defer gdb.Close()

	tagmgr = newTagManager(gdb)
	postmgr = newPostManager(gdb)
	optmgr = newOptionsModel(gdb)
	auther = &GenericAuth{}
	auther.SetLogin(optmgr.GetDef("login", "x"))
	auther.SetKey(config.key)
	uploadmgr = NewFileUpload(file_managers.NewLocalFileManager(config.files))
	backupmgr = NewBlogBackup(gdb)
	cmtmgr = newCommentManager(gdb)
	postcmtsmgr = newPostCommentsManager(gdb)
	fileredir = NewFileRedirect(config.base, config.files, config.fileHost)

	gin.DisableConsoleColor()
	router := gin.Default()

	tagapi := router.Group("/tags")

	tagapi.GET("/list", func(c *gin.Context) {

	})

	optapi := router.Group("/options")

	optapi.GET("/has", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		name := c.DefaultQuery("name", "")
		err := optmgr.Has(name)
		has := name != "" && err == nil
		EndReq(c, err, has)
	})

	optapi.GET("/get", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		name := c.DefaultQuery("name", "")
		val, err := optmgr.Get(name)
		EndReq(c, err, val)
	})

	optapi.POST("/set", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		name := c.DefaultPostForm("name", "")
		val := c.DefaultPostForm("value", "")
		err := optmgr.Set(name, val)
		EndReq(c, err, val)
	})

	optapi.POST("/del", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		name := c.DefaultPostForm("name", "")
		err := optmgr.Del(name)
		EndReq(c, err, nil)
	})

	optapi.GET("/list", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		items, err := optmgr.List()
		EndReq(c, err, items)
	})

	postapi := router.Group("/posts")

	postapi.POST("/update", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		var err error
		pidstr, has := c.GetPostForm("pid")
		pid, err := strconv.ParseInt(pidstr, 10, 64)
		if !has || err != nil || pid < 0 {
			c.String(400, "expect: pid")
			return
		}
		typ, has := c.GetPostForm("source_type")
		if !has {
			c.String(400, "expect: source_type")
			return
		}
		source, has := c.GetPostForm("source")
		if !has {
			c.String(400, "expect: source")
			return
		}

		err = postmgr.update(pid, typ, source)
		EndReq(c, err, nil)
	})

	backupapi := router.Group("/backups")

	backupapi.GET("backup", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}

		var sb bytes.Buffer
		err := backupmgr.Backup(&sb)
		EndReq(c, err, sb.String())
	})

	routerV1(router)
	routerInternalV1(router)

	router.Run(config.listen)
}

func toInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func routerInternalV1(router *gin.Engine) {
	v1 := router.Group("/.v1/")

	v1.POST("/send_mail", func(c *gin.Context) {
		author := c.PostForm("author")
		email := c.PostForm("email")
		subject := c.PostForm("subject")
		body := c.PostForm("body")

		log.Println("send_mail:", author, email, subject, body)

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
			if err = mc.AddTo(author, email); err != nil {
				log.Println("AddTo:", email, err)
				return
			}
			if err = mc.Send(subject, body); err != nil {
				log.Println(err)
				return
			}
		}()
	})
}

func routerV1(router *gin.Engine) {
	v1 := router.Group("/v1")

	posts := v1.Group("/posts")

	posts.GET("", func(c *gin.Context) {
		rets, err := getAllPosts(gdb)
		EndReq(c, err, rets)
	})

	posts.GET("/:parent/files/*name", func(c *gin.Context) {
		referrer := strings.ToLower(c.GetHeader("referer"))
		if strings.Contains(referrer, "://blog.csdn.net") {
			c.Redirect(302, "/1.jpg")
			return
		}
		parent := toInt64(c.Param("parent"))
		name := c.Param("name")
		if strings.Contains(name, "/../") {
			c.String(400, "bad file")
			return
		}
		logged := auth(c, false)
		path := fileredir.Redirect(logged, fmt.Sprintf("%d/%s", parent, name))
		c.Redirect(302, path)
	})

	posts.GET("/:parent/comments:count", func(c *gin.Context) {
		parent := toInt64(c.Param("parent"))
		count := postmgr.getCommentCount(parent)
		EndReq(c, true, count)
	})

	posts.GET("/:parent/comments", func(c *gin.Context) {
		var err error

		parent := toInt64(c.Param("parent"))
		offset := toInt64(c.Query("offset"))
		count := toInt64(c.Query("count"))
		order := c.DefaultQuery("order", "asc")

		cmts, err := postcmtsmgr.GetPostComments(0, offset, count, parent, order == "asc")

		if err != nil {
			EndReq(c, err, nil)
			return
		}

		var loggedin = auth(c, false)

		for _, c := range cmts {
			c.private = loggedin
		}

		EndReq(c, true, cmts)
	})

	posts.POST("/:parent/comments", func(c *gin.Context) {
		var err error
		var cmt Comment
		var loggedin bool

		loggedin = auth(c, false)

		cmt.PostID = toInt64(c.Param("parent"))
		cmt.Parent = toInt64(c.DefaultPostForm("parent", "0"))
		cmt.Author = c.DefaultPostForm("author", "")
		cmt.EMail = c.DefaultPostForm("email", "")
		cmt.URL = c.DefaultPostForm("url", "")
		cmt.IP = c.ClientIP()
		cmt.Date = datetime.MyLocal()
		cmt.Content = c.DefaultPostForm("content", "")

		if err = postmgr.has(cmt.PostID); err != nil {
			log.Println("找不到文章")
			EndReq(c, err, nil)
			return
		}

		if !loggedin {
			{
				notAllowedEmails := strings.Split(optmgr.GetDef("not_allowed_emails", ""), ",")
				if adminEmail := optmgr.GetDef("email", ""); adminEmail != "" {
					notAllowedEmails = append(notAllowedEmails, adminEmail)
				}

				log.Println(notAllowedEmails)

				// TODO use regexp to detect equality.
				for _, email := range notAllowedEmails {
					if email != "" && cmt.EMail != "" && strings.EqualFold(email, cmt.EMail) {
						EndReq(c, errors.New("不能使用此邮箱地址"), nil)
						return
					}
				}
			}
			{
				notAllowedAuthors := strings.Split(optmgr.GetDef("not_allowed_authors", ""), ",")
				if adminName := optmgr.GetDef("nickname", ""); adminName != "" {
					notAllowedAuthors = append(notAllowedAuthors, adminName)
				}

				for _, author := range notAllowedAuthors {
					if author != "" && cmt.Author != "" && strings.EqualFold(author, cmt.Author) {
						EndReq(c, errors.New("不能使用此昵称"), nil)
						return
					}
				}
			}
		}

		if err = cmtmgr.CreateComment(&cmt); err != nil {
			EndReq(c, err, nil)
			return
		}

		retCmt := c.DefaultQuery("return_cmt", "0") == "1"

		if !retCmt {
			EndReq(c, nil, gin.H{
				"id": cmt.ID,
			})
		} else {
			cmts, err := postcmtsmgr.GetPostComments(cmt.ID, 0, 1, cmt.PostID, true)
			if err != nil || len(cmts) == 0 {
				EndReq(c, errors.New("error get comment"), nil)
				return
			}
			cmts[0].private = !loggedin
			EndReq(c, err, cmts[0])
		}

		// TODO send email
	})

	posts.DELETE("/:parent/comments/:name", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}

		var err error

		parent := toInt64(c.Param("parent"))
		id := toInt64(c.Param("name"))

		// TODO check referrer
		_ = parent

		err = postcmtsmgr.DeletePostComment(id)
		EndReq(c, err, nil)
	})

	posts.GET("/:parent/files", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}

		files, err := uploadmgr.List(c)
		EndReq(c, err, files)
	})

	posts.POST("/:parent/files", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}

		err := uploadmgr.Upload(c)
		EndReq(c, err, nil)
	})

	posts.DELETE("/:parent/files/*name", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}

		err := uploadmgr.Delete(c)
		EndReq(c, err, nil)
	})

	archives := v1.Group("/archives")

	archives.GET("/categories/:name", func(c *gin.Context) {
		id := toInt64(c.Param("name"))
		ps, err := postmgr.GetPostsByCategory(id)
		EndReq(c, err, ps)
	})

	archives.GET("/tags/:name", func(c *gin.Context) {
		tag := c.Param("name")
		ps, err := postmgr.GetPostsByTags(tag)
		EndReq(c, err, ps)
	})

	archives.GET("/dates/:year/:month", func(c *gin.Context) {
		year := toInt64(c.Param("year"))
		month := toInt64(c.Param("month"))

		ps, err := postmgr.GetPostsByDate(year, month)
		EndReq(c, err, ps)
	})

	tools := v1.Group("/tools")

	tools.POST("/aes2htm", func(c *gin.Context) {
		aes2htm(c)
	})

	v1.Group("/sitemap.xml").GET("", func(c *gin.Context) {
		host := "https://" + optmgr.GetDef("home", "localhost")
		maps, err := createSitemap(gdb, host)
		if err != nil {
			EndReq(c, err, nil)
			return
		}
		c.Header("Content-Type", "application/xml")
		c.String(200, "%s", maps)
	})
}
