package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
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
	Tags_prev 			[]uint8	`json:"tags_prev"`
	Tags_prev2 			uint	
	Name 				string
	Slug 				string
	Count 				uint
}

func start_base() (db_mySql *sql.DB, db_postgres *sql.DB){
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

func all_info(db_mySql *sql.DB, db_postgres *sql.DB) {
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
	posts_id, err := db_postgres.Query("select old_id, tags_id from posts")
	if err != nil {
		panic(err)
	}
	for posts_id.Next() {
		var tags Tags
		err = posts_id.Scan(&tags.Post_id, &tags.Tags_prev)
		if err != nil {
			panic(err)
		}

		wp_term_relationships, err := db_mySql.Query("select term_taxonomy_id from wp_term_relationships where object_id = ?", tags.Post_id)
		if err != nil {
			panic(err)
		}
		
		for wp_term_relationships.Next() {
			err = wp_term_relationships.Scan(&tags.Term_taxonomy_id)
			if err != nil {
				panic(err)
			}

			wp_term_taxonomy, err := db_mySql.Query("select term_id, count from wp_term_taxonomy where term_taxonomy_id = ? and taxonomy != 'yst_prominent_words' and taxonomy != 'amp_validation_error';", tags.Term_taxonomy_id)
			if err != nil {
				panic(err)
			}

			for wp_term_taxonomy.Next() {
				err = wp_term_taxonomy.Scan(&tags.Tag_id, &tags.Count)
				if err != nil {
					panic(err)
				}

				wp_terms, err := db_mySql.Query("select name, slug from wp_terms where term_id = ?;", tags.Tag_id)
				if err != nil {
					panic(err)
				}

				for wp_terms.Next() {
					err = wp_terms.Scan(&tags.Name, &tags.Slug)
					if err != nil {
						panic(err)
					}
					err = db_postgres.QueryRow("insert into tags (name, slug, count) values ($1, $2, $3) returning id;", tags.Name, tags.Slug, tags.Count).Scan(&tags.Tags_prev2)
					if err != nil {
						panic(err)
					}

					res := uint8(tags.Tags_prev2)
					tags.Tags_prev = append(tags.Tags_prev, res)
					db_postgres.QueryRow("update posts set tags_id = $1 where old_id = $2", tags.Tags_prev, tags.Post_id)
				}
				wp_terms.Close()
			}
			wp_term_taxonomy.Close()
		}
		wp_term_relationships.Close()
	}
	posts_id.Close()
}

func main() {
		//Подключаем базы
	db_mySql, db_postgres := start_base()
		//Перенос основной инфы Post
	//all_info(db_mySql, db_postgres)
		//Все изображения
	//img(db_mySql, db_postgres)
		//Все теги
	tags(db_mySql, db_postgres)
		//Закрытие обращений к базе
	db_postgres.Close()
	db_mySql.Close()
}

