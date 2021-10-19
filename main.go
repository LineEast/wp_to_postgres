package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

type Post struct {
	Id uint `json:"id"`
	Author uint `json:"author"`
	Date string `json:"date"`
	Content string `json:"content"`
	Content_short string `json:"Content_short"`
	Title string `json:"title"`
	Img string `json:"img"`
	Tags_id []uint `json:"tags_id"`
}

func startBase() (db_mySql *sql.DB, db_postgres *sql.DB){
	db_mySql, err := sql.Open("mysql", "root:1337228@tcp(127.0.0.1:3306)/nuancesprog")
	if err != nil {
		panic(err)
	}
	db_postgres, err = sql.Open("postgres", "user=line_east dbname=line_east sslmode=disable")
	if err != nil {
		fmt.Println(err)
	}
	return
}

func allInfo(db_mySql *sql.DB, db_postgres *sql.DB) {
	mysql_select, err := db_mySql.Query("select id, post_author, post_date, post_content, post_excerpt, post_title from wp_posts where post_status = 'publish' and ping_status = 'open' and post_type != 'revision';")
	if err != nil {
		panic(err)
	}

	for mysql_select.Next() {
		var post Post
		err = mysql_select.Scan(&post.Id, &post.Author, &post.Date, &post.Content, &post.Content_short, &post.Title)
		if err != nil {
			panic(err)
		}
		_, err := db_postgres.Exec("insert into posts (old_id, author, post, content, Content_short, title) values ($1, $2, $3, $4, $5, $6)", post.Id, post.Author, post.Date, post.Content, post.Content_short, post.Title)
		if err != nil {
			panic(err)
		}
	}

	mysql_select.Close()
}

func img(db_mySql *sql.DB, db_postgres *sql.DB) {
	mysql_select, err := db_mySql.Query("select post_parent, guid from wp_posts where post_type = 'attachment';")
	if err != nil {
		panic(err)
	}
	for mysql_select.Next() {
		var post Post
		err = mysql_select.Scan(&post.Id, &post.Img)
		if err != nil {
			panic(err)
		}
		_, err := db_postgres.Exec("update posts set img = $1 where old_id = $2;", post.Img, post.Id)
		if err != nil {
			panic(err)
		}
	}

	mysql_select.Close()
}

func tags(db_mySql *sql.DB, db_postgres *sql.DB) {
	
}

func main() {
	//Подключаем базы
	db_mySql, db_postgres := startBase()
	//Перенос основной инфы Post
	allInfo(db_mySql, db_postgres)
	//Все изображения
	img(db_mySql, db_postgres)
	//Все теги
	tags(db_mySql, db_postgres)
	//Закрытие всех обращений к базе
	db_postgres.Close()
	db_mySql.Close()
}