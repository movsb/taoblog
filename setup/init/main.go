package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	_ "github.com/go-sql-driver/mysql"

	"github.com/movsb/taoblog/modules/stdinlinereader"
	"github.com/movsb/taoblog/service/models"
	"github.com/movsb/taorm/taorm"
)

const dbVer = 12

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
	createDatabaseUser(db, blogDatabaseName, blogDatabaseUserName, blogDatabasePassword)
}

func createDatabase(db *taorm.DB, dbName string) {
	query := fmt.Sprintf(`CREATE DATABASE %s`, dbName)
	db.MustExec(query)
}

func createDatabaseUser(db *taorm.DB, database string, username string, password string) {
	query := fmt.Sprintf(
		`CREATE USER "%s"@"%s" IDENTIFIED BY "%s"`,
		username, "%", password,
	)
	db.MustExec(query)
	query = fmt.Sprintf(
		`GRANT ALL ON %s.* TO "%s"@"%s";`,
		database, username, "%",
	)
	db.MustExec(query)
}

func createDatabaseTables(db *taorm.DB, dbName string) {
	queryBytes, err := ioutil.ReadFile("data/schemas.sql")
	if err != nil {
		panic(err)
	}
	query := fmt.Sprintf("USE %s;", dbName)
	query += string(queryBytes)
	db.MustExec(query)

	query = fmt.Sprintf("INSERT INTO options (name,value) VALUES (?,?)")
	db.MustExec(query, "db_ver", dbVer)

	rootCategory := models.Category{
		ID:   1,
		Name: "未分类",
		Slug: "uncategorized",
		Path: "/",
	}
	db.Model(&rootCategory).MustCreate()
}
