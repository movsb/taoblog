package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type xConfig struct {
	listen   string
	username string
	password string
	database string
	key      string
	files    string
}

var gkey string
var config xConfig
var gdb *sql.DB
var tagmgr *xTagManager
var postmgr *xPostManager
var optmgr *xOptionsModel
var auther *GenericAuth
var uploadmgr *FileUpload

type xJSONRet struct {
	Code int         `json:"code"`
	Msgs string      `json:"msgs"`
	Data interface{} `json:"data"`
}

func auth(c *gin.Context) bool {
	if auther.AuthHeader(c) || auther.AuthCookie(c) {
		return true
	}
	finishError(c, -1, errors.New("auth error"))
	return false
}

func main() {
	flag.StringVar(&config.listen, "listen", "127.0.0.1:2564", "the port to which the server listen")
	flag.StringVar(&config.username, "username", "taoblog", "the database username")
	flag.StringVar(&config.password, "password", "taoblog", "the database password")
	flag.StringVar(&config.database, "database", "taoblog", "the database name")
	flag.StringVar(&config.key, "key", "", "api key")
	flag.StringVar(&config.files, "files", ".", "the files folder")
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
	uploadmgr = NewFileUpload(config.files)

	gin.DisableConsoleColor()
	router := gin.Default()

	router.GET("/all-posts.go", func(c *gin.Context) {
		rets, err := getAllPosts(gdb)
		if err != nil {
			c.JSON(500, fmt.Sprint(err))
			return
		}

		c.JSON(200, rets)
	})

	tagapi := router.Group("/tags")

	tagapi.GET("/list", func(c *gin.Context) {

	})

	optapi := router.Group("/option")

	optapi.GET("/has", func(c *gin.Context) {
		if !auth(c) {
			return
		}
		name := c.DefaultQuery("name", "")
		err := optmgr.Has(name)
		has := name != "" && err == nil
		finishDone(c, 0, "", has)
	})

	optapi.GET("/get", func(c *gin.Context) {
		if !auth(c) {
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
		if !auth(c) {
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
		if !auth(c) {
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
		if !auth(c) {
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
		if !auth(c) {
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

	postapi.GET("/comment-count", func(c *gin.Context) {
		var err error
		pidstr, has := c.GetQuery("pid")
		pid, err := strconv.ParseInt(pidstr, 10, 64)
		if !has || err != nil || pid < 0 {
			c.String(400, "expect: pid")
			return
		}
		count := postmgr.getCommentCount(pid)
		finishDone(c, 0, "", count)
	})

	uploadapi := router.Group("/upload")

	uploadapi.POST("/upload", func(c *gin.Context) {
		if !auth(c) {
			return
		}

		if err := uploadmgr.Upload(c); err != nil {
			finishError(c, -1, err)
			return
		}

		files := uploadmgr.List(c)
		finishDone(c, 0, "", files)
	})

	uploadapi.GET("/list", func(c *gin.Context) {
		if !auth(c) {
			return
		}

		files := uploadmgr.List(c)
		finishDone(c, 0, "", files)
	})

	uploadapi.POST("/delete", func(c *gin.Context) {
		if !auth(c) {
			return
		}

		ok := uploadmgr.Delete(c)
		if ok {
			finishDone(c, 0, "", ok)
		} else {
			finishError(c, -1, nil)
		}
	})

	router.Run(config.listen)
}
