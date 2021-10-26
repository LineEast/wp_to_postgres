package main

import (
	"database/sql"
	"context"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Post struct {
	Id					uint	`json:"id"`
	Author				uint 	`json:"author"`
	Date 				string 	`json:"date"`
	Content 			string 	`json:"content"`
	Content_short 		string 	`json:"Content_short"`
	Title 				string 	`json:"title"`
	Img 				string 	`json:"img"`
	Tags_id 			[]uint 	`json:"tags_id"`
}

type Tags struct {
	Tag_id 				uint 	`json:"id"`
	Post_id 			uint 	`json:"post_id"`
	Term_taxonomy_id 	uint
	Tags_prev 			[]uint	`json:"tags_prev"`
	Tags_prev2 			uint	
	Name 				string
	Slug 				string
	Count 				uint
	Taxonomy 			string

	New_id				uint
}

func start_base() (db_mySql *sql.DB, db_postgres *pgxpool.Pool){
	db_mySql, err := sql.Open("mysql", "root:1337228@tcp(127.0.0.1:3306)/nuancesprog")
	error_short(err)
	db_postgres, err = pgxpool.Connect(context.Background(), "postgres://line_east@localhost:5432/line_east")
	error_short(err)
	return
}

func all_info(db_mySql *sql.DB, db_postgres *pgxpool.Pool) {
	mysql_select, err := db_mySql.Query("select id, post_author, post_date, post_content, post_excerpt, post_title from wp_posts where post_status = 'publish' and ping_status = 'open' and post_type != 'revision';")
	error_short(err)

	for mysql_select.Next() {
		var post Post
		err = mysql_select.Scan(&post.Id, &post.Author, &post.Date, &post.Content, &post.Content_short, &post.Title)
		error_short(err)
		_, err := db_postgres.Exec(context.Background(), "insert into posts (old_id, author, post, content, Content_short, title) values ($1, $2, $3, $4, $5, $6)", post.Id, post.Author, post.Date, post.Content, post.Content_short, post.Title)
		error_short(err)
	}

	mysql_select.Close()
}

func img(db_mySql *sql.DB, db_postgres *pgxpool.Pool) {
	mysql_select, err := db_mySql.Query("select post_parent, guid from wp_posts where post_type = 'attachment';")
	error_short(err)

	for mysql_select.Next() {
		var post Post
		err = mysql_select.Scan(&post.Id, &post.Img)
		error_short(err)
		_, err := db_postgres.Exec(context.Background(), "update posts set img = $1 where old_id = $2;", post.Img, post.Id)
		error_short(err)
	}
	mysql_select.Close()
}

func tags(db_mySql *sql.DB, db_postgres *pgxpool.Pool) {
	post_id, err := db_postgres.Query(context.Background(), "select old_id, tags_id from posts;")
	error_short(err)

	for post_id.Next() {
		var tags Tags
		err = post_id.Scan(&tags.Post_id, &tags.Tags_prev)
		error_short(err)

		wp_term_relationships, err := db_mySql.Query("select term_taxonomy_id from wp_term_relationships where object_id = ?;", tags.Post_id)
		error_short(err)

		for wp_term_relationships.Next() {
			count := 0

			err = wp_term_relationships.Scan(&tags.Term_taxonomy_id)
			error_short(err)
			
			err = db_mySql.QueryRow("select term_id, count, taxonomy from wp_term_taxonomy where term_taxonomy_id = ?;", tags.Term_taxonomy_id).Scan(&tags.Tag_id, &tags.Count, &tags.Taxonomy)
			error_short(err)
			if (tags.Taxonomy == "post_tag" || tags.Taxonomy == "category") {
				err = db_mySql.QueryRow("select name, slug from wp_terms where term_id = ?", tags.Tag_id).Scan(&tags.Name, &tags.Slug)
				error_short(err)

				err := db_postgres.QueryRow(context.Background(), "select count(*) from tags where name = $1", tags.Name).Scan(&count)
				error_short(err)

				if count == 0 {
					err = db_postgres.QueryRow(context.Background(), "insert into tags (name, slug, count, taxonomy) values ($1, $2, $3, $4) returning id;", tags.Name, tags.Slug, tags.Count, tags.Taxonomy).Scan(&tags.Tags_prev2)
					error_short(err)

					tags.Tags_prev = append(tags.Tags_prev, tags.Tags_prev2)

					db_postgres.QueryRow(context.Background(), "update posts set tags_id = $1 where old_id = $2", tags.Tags_prev, tags.Post_id)
				} else {
					err =  db_postgres.QueryRow(context.Background(), "select id from tags where name = $1", tags.Name).Scan(&tags.Tags_prev2)
					error_short(err)

					tags.Tags_prev = append(tags.Tags_prev, tags.Tags_prev2)

					err = db_postgres.QueryRow(context.Background(), "update posts set tags_id = $1 where old_id = $2 returning id", tags.Tags_prev, tags.Post_id).Scan(&tags.New_id)
					error_short(err)

					fmt.Println(tags.New_id)
				}
			}
		}
		wp_term_relationships.Close()
	}
	post_id.Close()
}

func error_short(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
		//Подключаем базы
	fmt.Println("start_base: start")
	db_mySql, db_postgres := start_base()
	fmt.Println("start_base: end")
		//Перенос основной инфы Post
	// fmt.Println("all_info: start")
	// all_info(db_mySql, db_postgres)
	// fmt.Println("all_info: end")
	//  	//Все изображения
	// fmt.Println("img: start")
	// img(db_mySql, db_postgres)
	// fmt.Println("img: end")
		//Все теги
	fmt.Println("tags: start")
	tags(db_mySql, db_postgres)
	fmt.Println("tags: start")
		//Закрытие обращений к базе
	db_postgres.Close()
	db_mySql.Close()
}