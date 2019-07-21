package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	_ "github.com/go-sql-driver/mysql"

	"github.com/movsb/taoblog/auth"
	"github.com/movsb/taoblog/modules/stdinlinereader"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm/taorm"
)

const dbVer = 10

var liner = stdinlinereader.NewStdinLineReader()

func main() {
	mysqlAdminUsername := liner.MustReadLine("MySQL数据库管理员用户名：")
	mysqlAdminPassword := liner.MustReadLine("MySQL数据库管理员密码：")
	dataSource := fmt.Sprintf("%s:%s@/?multiStatements=true", mysqlAdminUsername, mysqlAdminPassword)
	rdb, err := sql.Open("mysql", dataSource)
	if err != nil {
		panic(err)
	}
	if err := rdb.Ping(); err != nil {
		panic(err)
	}
	defer rdb.Close()

	blogDatabaseName := liner.MustReadLine("博客数据库名字：")
	blogDatabaseUserName := liner.MustReadLine("博客数据库用户名：")
	blogDatabasePassword := liner.MustReadLine("博客数据库密码：")

	db := taorm.NewDB(rdb)
	createDatabase(db, blogDatabaseName)
	createDatabaseTables(db, blogDatabaseName)
	createDatabaseUser(db, blogDatabaseUserName, blogDatabasePassword)
	createBlogUser(db)
	createBlogInfo(db)
}

func createDatabase(db *taorm.DB, dbName string) {
	query := fmt.Sprintf(`CREATE DATABASE %s`, dbName)
	db.MustExec(query)
}

func createDatabaseUser(db *taorm.DB, username string, password string) {
	query := fmt.Sprintf(
		`CREATE USER "%s"@"%s" IDENTIFIED BY "%s"`,
		username, "localhost", password,
	)
	db.MustExec(query)
	query = fmt.Sprintf(
		`GRANT ALL ON %s.* TO "%s"@"%s";`,
		username, password, "localhost",
	)
	db.MustExec(query)
}

func createDatabaseTables(db *taorm.DB, dbName string) {
	queryBytes, err := ioutil.ReadFile("../data/schemas.sql")
	if err != nil {
		panic(err)
	}
	query := fmt.Sprintf("USE %s;", dbName)
	query += string(queryBytes)
	db.MustExec(query)

	query = fmt.Sprintf("INSERT INTO options (name,value) VALUES (?,?)")
	db.MustExec(query, "db_ver", dbVer)
	db.MustExec(query, "home", "localhost")

	rootCategory := models.Category{
		ID:   1,
		Name: "未分类",
		Slug: "uncategorized",
		Path: "/",
	}
	db.Model(&rootCategory, "categories").MustCreate()
}

func createBlogUser(db *taorm.DB) {
	blogUsername := liner.MustReadLine("博客用户名：")
	blogPassword := liner.MustReadLine("博客密码：")
	googleClientID := liner.MustReadLine("谷歌ClientID：")
	adminGoogleID := liner.MustReadLine("管理员谷歌ID：")
	savedAuth := auth.SavedAuth{
		Username:       blogUsername,
		Password:       auth.HashPassword(blogPassword),
		GoogleClientID: googleClientID,
		AdminGoogleID:  adminGoogleID,
	}
	query := fmt.Sprintf("INSERT INTO options (name,value) VALUES (?,?)")
	db.MustExec(query, "login", savedAuth.Encode())
}

func createBlogInfo(db *taorm.DB) {
	blogName := liner.MustReadLine("博客名字：")
	blogDesc := liner.MustReadLine("博客描述：")
	query := fmt.Sprintf("INSERT INTO options (name,value) VALUES (?,?)")
	db.MustExec(query, "blog_name", blogName)
	query = fmt.Sprintf("INSERT INTO options (name,value) VALUES (?,?)")
	db.MustExec(query, "blog_desc", blogDesc)
}
