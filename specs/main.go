package main

import (
	"database/sql"
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
}

var config xConfig
var gdb *sql.DB
var tagmgr *xTagManager

func main() {
	flag.StringVar(&config.listen, "listen", "127.0.0.1:2564", "the port to which the server listen")
	flag.StringVar(&config.username, "username", "taoblog", "the database username")
	flag.StringVar(&config.password, "password", "taoblog", "the database password")
	flag.StringVar(&config.database, "database", "taoblog", "the database name")
	flag.Parse()

	var err error
	dataSource := fmt.Sprintf("%s:%s@/%s", config.username, config.password, config.database)
	gdb, err = sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}

	gdb.SetMaxIdleConns(10)

	defer gdb.Close()

	tagmgr = newTagManager(gdb)

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

	postapi := router.Group("/posts")

	postapi.GET("/get-tag-names", func(c *gin.Context) {
		pidstr, has := c.GetQuery("pid")
		pid, err := strconv.ParseInt(pidstr, 10, 64)
		if !has || err != nil || pid < 0 {
			c.String(400, "invalid argument: pid")
			return
		}
		names := tagmgr.getTagNames(pid)
		c.JSON(200, names)
	})

	router.Run(config.listen)
}
