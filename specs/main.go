package main

import (
	"database/sql"
	"flag"
	"fmt"

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

	defer gdb.Close()

	gin.DisableConsoleColor()
	router := gin.Default()

	router.GET("/all-posts.go", func(c *gin.Context) {
		rets, err := getAllPosts(gdb)
		if err != nil {
			c.JSON(500, err)
			return
		}

		c.JSON(200, rets)
	})

	router.Run(config.listen)
}
