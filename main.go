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
	Title string `json:"title"`
	Img string `json:"img"`
	Tags_id []uint `json:"tags_id"`
}

func main() {
	//Подключаем базы
	db_mySql, err := sql.Open("mysql", "root:1337228@tcp(127.0.0.1:3306)/nuancesprog")
	if err != nil {
		panic(err)
	}
	db_postgres, err := sql.Open("postgres", "user=line_east password=1337228 dbname=nuncesprog_new sslmod=disable")
	if err != nil {
		panic(err)
	}

	//Обращение в исходник
	mysql_select, err := db_mySql.Query("select id, post_author, post_date, post_content, post_title from wp_posts where post_status = 'publish' and ping_status = 'open' and post_type != 'revision';")
	if err != nil {
		panic(err)
	}

	for mysql_select.Next() {
		var post Post
		err = mysql_select.Scan(&post.Id, &post.Author, &post.Date, &post.Content, &post.Title)
		if err != nil {
			panic(err)
		}

		//post.Content = removeComments(post.Content)

		//
		//fmt.Println(fmt.Sprintf("old_id: %d, author_id: %d, date: %s", post.Id, post.Author, post.Date))
	}

	//Закрытие всех обращений к базе
	mysql_select.Close()

	db_postgres.Close()
	db_mySql.Close()
}
//Обратиться в базу