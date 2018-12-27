package main

import (
	"database/sql"
	"flag"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/front"
	"github.com/movsb/taoblog/gateway"
	"github.com/movsb/taoblog/modules/utils"
	"github.com/movsb/taoblog/service"
	"github.com/movsb/taoblog/service/modules/file_managers"
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
var uploadmgr *FileUpload
var fileredir *FileRedirect
var theAuth *auth.Auth

var theFront *front.Front

//var theAdmin *admin.Admin
var implServer *service.ImplServer
var theGateway *gateway.Gateway

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

	theAuth = &auth.Auth{}
	implServer = service.NewImplServer(gdb, theAuth)
	theAuth.SetLogin(implServer.GetDefaultStringOption("login", "x"))
	theAuth.SetKey(config.key)
	uploadmgr = NewFileUpload(file_managers.NewLocalFileManager(config.files))
	fileredir = NewFileRedirect(config.base, config.files, config.fileHost)

	router := gin.Default()

	v2 := router.Group("/v2")

	//theAdmin = admin.NewAdmin(implServer, &router.RouterGroup)
	theFront = front.NewFront(implServer, theAuth, &router.RouterGroup, v2)

	//routerV1(router)

	v2.Use(func(c *gin.Context) {
		defer func() {
			if e := recover(); e != nil {
				if err, ok := e.(error); ok {
					if err == sql.ErrNoRows {
						c.Status(404)
						return
					}
				}
				panic(e)
			}
		}()
		c.Next()
	})
	theGateway = gateway.NewGateway(v2, implServer, theAuth)

	router.Run(config.listen)
}

func routerV1(router *gin.Engine) {
	v1 := router.Group("/v1")

	v1.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	posts := v1.Group("/posts")

	posts.GET("/:parent/files/*name", func(c *gin.Context) {
		referrer := strings.ToLower(c.GetHeader("referer"))
		if strings.Contains(referrer, "://blog.csdn.net") {
			c.Redirect(302, "/1.jpg")
			return
		}
		parent := utils.MustToInt64(c.Param("parent"))
		name := c.Param("name")
		if strings.Contains(name, "/../") {
			c.String(400, "bad file")
			return
		}
		logged := false
		path := fileredir.Redirect(logged, fmt.Sprintf("%d/%s", parent, name))
		c.Redirect(302, path)
	})

	/*
		posts.GET("/:parent/files", theAuth.Middle, func(c *gin.Context) {
			files, err := uploadmgr.List(c)
			EndReq(c, err, files)
		})

		posts.POST("/:parent/files/:name", theAuth.Middle, func(c *gin.Context) {
			err := uploadmgr.Upload(c)
			EndReq(c, err, nil)
		})

		posts.DELETE("/:parent/files/:name", theAuth.Middle, func(c *gin.Context) {
			err := uploadmgr.Delete(c)
			EndReq(c, err, nil)
		})
	*/

	/*
		v1.GET("/posts!all", func(c *gin.Context) {
			var posts []*PostForArchiveQuery
			if p, ok := memcch.Get("posts:all"); ok {
				posts = p.([]*PostForArchiveQuery)
			} else {
				p, err := postmgr.ListAllPosts(gdb)
				if err != nil {
					EndReq(c, err, posts)
					return
				}
				memcch.Set("posts:all", p)
				posts = p
			}
			EndReq(c, nil, posts)
		})

		archives := v1.Group("/archives")

		archives.GET("/categories/:name", func(c *gin.Context) {
			id := toInt64(c.Param("name"))
			ps, err := postmgr.GetPostsByCategory(gdb, id)
			EndReq(c, err, ps)
		})

		archives.GET("/dates/:year/:month", func(c *gin.Context) {
			year := toInt64(c.Param("year"))
			month := toInt64(c.Param("month"))

			ps, err := postmgr.GetPostsByDate(gdb, year, month)
			EndReq(c, err, ps)
		})

		tools := v1.Group("/tools")

		tools.POST("/aes2htm", func(c *gin.Context) {
			aes2htm(c)
		})

	*/
}
