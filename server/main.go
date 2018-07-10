package main

import (
	"bytes"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"./internal/file_managers"
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

type xJSONRet struct {
	Code int         `json:"code"`
	Msgs string      `json:"msgs"`
	Data interface{} `json:"data"`
}

func auth(c *gin.Context, finish bool) bool {
	if auther.AuthHeader(c) || auther.AuthCookie(c) {
		return true
	}
	if finish {
		finishError(c, -1, errors.New("auth error"))
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
		finishDone(c, 0, "", has)
	})

	optapi.GET("/get", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		name := c.DefaultQuery("name", "")
		val, err := optmgr.Get(name)
		if err != nil {
			finishError(c, -1, err)
		} else {
			finishDone(c, 0, "", val)
		}
	})

	optapi.POST("/set", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		name := c.DefaultPostForm("name", "")
		val := c.DefaultPostForm("value", "")
		if err := optmgr.Set(name, val); err == nil {
			finishDone(c, 0, "", nil)
		} else {
			finishError(c, -1, err)
		}
	})

	optapi.POST("/del", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		name := c.DefaultPostForm("name", "")
		if err := optmgr.Del(name); err == nil {
			finishDone(c, 0, "", nil)
		} else {
			finishError(c, -1, err)
		}
	})

	optapi.GET("/list", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}
		items, err := optmgr.List()
		if err != nil {
			finishError(c, -1, err)
		}

		finishDone(c, 0, "", items)
	})

	postapi := router.Group("/posts")

	postapi.GET("/has", func(c *gin.Context) {
		pidstr, has := c.GetQuery("pid")
		pid, err := strconv.ParseInt(pidstr, 10, 64)
		if !has || err != nil || pid < 0 {
			c.String(400, "invalid argument: pid")
			return
		}
		if err := postmgr.has(pid); err != nil {
			finishError(c, -1, err)
		} else {
			finishDone(c, 0, "", nil)
		}
	})

	postapi.GET("/get-tag-names", func(c *gin.Context) {
		pidstr, has := c.GetQuery("pid")
		pid, err := strconv.ParseInt(pidstr, 10, 64)
		if !has || err != nil || pid < 0 {
			c.String(400, "invalid argument: pid")
			return
		}
		if err := postmgr.has(pid); err != nil {
			finishError(c, -1, err)
			return
		}
		names := tagmgr.getTagNames(pid)
		finishDone(c, 0, "", names)
	})

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
		if err != nil {
			finishError(c, -1, err)
			return
		}

		finishDone(c, 0, "", nil)
	})

	backupapi := router.Group("/backups")

	backupapi.GET("backup", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}

		var sb bytes.Buffer
		var er error
		er = backupmgr.Backup(&sb)
		if er != nil {
			finishError(c, -1, er)
			return
		}
		finishDone(c, 0, "", sb.String())
	})

	routerV1(router)

	router.Run(config.listen)
}

func toInt64(s string) int64 {
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

func routerV1(router *gin.Engine) {
	v1 := router.Group("/v1")

	posts := v1.Group("/posts")

	posts.GET("", func(c *gin.Context) {
		rets, err := getAllPosts(gdb)
		if err != nil {
			c.JSON(500, fmt.Sprint(err))
			return
		}

		c.JSON(200, rets)
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
		finishDone(c, 0, "", count)
	})

	posts.GET("/:parent/comments", func(c *gin.Context) {
		var err error

		parent := toInt64(c.Param("parent"))
		offset := toInt64(c.Query("offset"))
		count := toInt64(c.Query("count"))
		order := c.DefaultQuery("order", "asc")

		cmts, err := postcmtsmgr.GetPostComments(offset, count, parent, order == "asc")

		if err != nil {
			finishError(c, -1, err)
			return
		}

		var loggedin = auth(c, false)

		for _, c := range cmts {
			c.private = loggedin
		}

		finishDone(c, 0, "", cmts)
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

		if err = postcmtsmgr.DeletePostComment(id); err != nil {
			finishError(c, -1, err)
			return
		}

		finishDone(c, 0, "", nil)
	})

	posts.GET("/:parent/files", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}

		if files, err := uploadmgr.List(c); err == nil {
			finishDone(c, 0, "", files)
		} else {
			finishError(c, -1, err)
		}
	})

	posts.POST("/:parent/files", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}

		if err := uploadmgr.Upload(c); err != nil {
			finishError(c, -1, err)
		} else {
			finishDone(c, 0, "", nil)
		}
	})

	posts.DELETE("/:parent/files/*name", func(c *gin.Context) {
		if !auth(c, true) {
			return
		}

		err := uploadmgr.Delete(c)
		if err == nil {
			finishDone(c, 0, "", nil)
		} else {
			finishError(c, -1, nil)
		}
	})

	archives := v1.Group("/archives")

	archives.GET("/categories/:name", func(c *gin.Context) {
		id := toInt64(c.Param("name"))
		ps, err := postmgr.GetPostsByCategory(id)
		if err != nil {
			finishError(c, -1, err)
			return
		}

		finishDone(c, 0, "", ps)
	})

	archives.GET("/tags/:name", func(c *gin.Context) {
		tag := c.Param("name")
		ps, err := postmgr.GetPostsByTags(tag)
		if err != nil {
			finishError(c, -1, err)
			return
		}

		finishDone(c, 0, "", ps)
	})

	archives.GET("/dates/:year/:month", func(c *gin.Context) {
		year := toInt64(c.Param("year"))
		month := toInt64(c.Param("month"))

		ps, err := postmgr.GetPostsByDate(year, month)
		if err != nil {
			finishError(c, -1, err)
			return
		}

		finishDone(c, 0, "", ps)
	})

	tools := v1.Group("/tools")

	tools.POST("/aes2htm", func(c *gin.Context) {
		aes2htm(c)
	})
}
