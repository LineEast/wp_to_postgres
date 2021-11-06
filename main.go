//go:build old

package main

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Post struct {
	OldID        uint   `json:"id"`
	Author       uint   `json:"author"`
	Date         string `json:"date"`
	Content      string `json:"content"`
	ContentShort string `json:"Content_short"`
	Title        string `json:"title"`
	Image        string `json:"img"`
	TagsID       []uint `json:"tags_id"`
}

type Tags struct {
	TagID          uint `json:"id"`
	PostID         uint `json:"post_id"`
	TermTaxonomyID uint
	TagsPrev       []uint `json:"tags_prev"`
	TagsPrev2      uint
	Name           string
	Slug           string
	Count          uint
	Taxonomy       string

	NewID uint
}

func startBase() (dbMySql *sql.DB, dbPostgres *pgxpool.Pool) {
	dbMySql, err := sql.Open("mysql", "root:#commit@tcp(127.0.0.1:3306)/nuancesprog")
	errorShort(err)
	dbPostgres, err = pgxpool.Connect(context.Background(), "postgres://postgres@localhost:5432/nuancesprog")
	errorShort(err)
	return
}

func allInfo(dbMySql *sql.DB, dbPostgres *pgxpool.Pool) {
	mysqlSelect, err := dbMySql.Query("select id, post_author, post_date, post_content, post_excerpt, post_title from wp_posts where post_status = 'publish' and ping_status = 'open' and post_type != 'revision' order by id;")
	errorShort(err)

	for mysqlSelect.Next() {
		var post Post
		err = mysqlSelect.Scan(&post.OldID, &post.Author, &post.Date, &post.Content, &post.ContentShort, &post.Title)
		errorShort(err)
		_, err := dbPostgres.Exec(context.Background(), "insert into posts (old_id, author, date, content, Content_short, title) values ($1, $2, $3, $4, $5, $6)", post.OldID, post.Author, post.Date, post.Content, post.ContentShort, post.Title)
		errorShort(err)
		fmt.Println(post.Date)
	}
	mysqlSelect.Close()
}

func image(dbMySql *sql.DB, dbPostgres *pgxpool.Pool) {
	post := Post{}
	mysqlSelect, err := dbMySql.Query("select wp_posts.guid as image from wp_postmeta left join wp_posts on wp_postmeta.meta_value = wp_posts.id where post_id = ? and meta_key = '_thumbnail_id';", post.OldID)
	errorShort(err)

	for mysqlSelect.Next() {
		post := Post{}
		err = mysqlSelect.Scan(&post.Image)
		errorShort(err)
		_, err := dbPostgres.Exec(context.Background(), "update posts set img = $1 where old_id = $2;", post.Image, post.OldID)
		errorShort(err)
	}
	mysqlSelect.Close()
}

func tags(dbMySql *sql.DB, dbPostgres *pgxpool.Pool) {
	count := 0

	postID, err := dbPostgres.Query(context.Background(), "select old_id, tags_id from posts order by old_id;")
	errorShort(err)

	for postID.Next() {
		var tags Tags
		err = postID.Scan(&tags.PostID, &tags.TagsPrev)
		errorShort(err)

		wpTermRelationships, err := dbMySql.Query("select term_taxonomy_id from wp_term_relationships where object_id = ?;", tags.PostID)
		errorShort(err)

		for wpTermRelationships.Next() {
			count = 0

			err = wpTermRelationships.Scan(&tags.TermTaxonomyID)
			errorShort(err)

			wpTermTaxonomySelect, err := dbMySql.Query("select term_id, count, taxonomy from wp_term_taxonomy where term_taxonomy_id = ?;", tags.TermTaxonomyID)
			errorShort(err)
			for wpTermTaxonomySelect.Next() {
				err = wpTermTaxonomySelect.Scan(&tags.TagID, &tags.Count, &tags.Taxonomy)
				errorShort(err)

				if tags.Taxonomy == "post_tag" || tags.Taxonomy == "category" {
					selectWpTerms, err := dbMySql.Query("select name, slug from wp_terms where term_id = ?;", tags.TagID)
					errorShort(err)
					for selectWpTerms.Next() {
						err = selectWpTerms.Scan(&tags.Name, &tags.Slug)
						errorShort(err)

						count = 0
						tagsCount, err := dbPostgres.Query(context.Background(), "select count(*) from tags where name = $1;", tags.Name)
						errorShort(err)

						for tagsCount.Next() {
							err = tagsCount.Scan(&count)
							errorShort(err)

							if count == 0 {
								tagsInsert, err := dbPostgres.Query(context.Background(), "insert into tags (name, slug, count, taxonomy) values ($1, $2, $3, $4) returning id;", tags.Name, tags.Slug, tags.Count, tags.Taxonomy)
								errorShort(err)

								for tagsInsert.Next() {
									err = tagsInsert.Scan(&tags.TagsPrev2)
									errorShort(err)

									tags.TagsPrev = append(tags.TagsPrev, tags.TagsPrev2)

									updatePosts, err := dbPostgres.Query(context.Background(), "update posts set tags_id = $1 where old_id = $2;", tags.TagsPrev, tags.PostID)
									errorShort(err)
									updatePosts.Close()
								}
								tagsInsert.Close()

							} else {
								selectTags, err := dbPostgres.Query(context.Background(), "select id from tags where name = $1;", tags.Name)
								errorShort(err)

								for selectTags.Next() {
									err = selectTags.Scan(&tags.TagsPrev2)
									errorShort(err)

									tags.TagsPrev = append(tags.TagsPrev, tags.TagsPrev2)
								}

								updatePosts, err := dbPostgres.Query(context.Background(), "update posts set tags_id = $1 where old_id = $2", tags.TagsPrev, tags.PostID)
								errorShort(err)
								selectTags.Close()
								updatePosts.Close()
							}
						}
						tagsCount.Close()
					}
					selectWpTerms.Close()
				}
			}
			wpTermTaxonomySelect.Close()
		}
		wpTermRelationships.Close()
	}
	postID.Close()
}

func errorShort(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	//Подключаем базы
	dbMySql, dbPostgres := startBase()
	//Перенос основной инфы Post
	allInfo(dbMySql, dbPostgres)
	//  	//Все изображения
	image(dbMySql, dbPostgres)
	//Все теги
	tags(dbMySql, dbPostgres)
	//Закрытие обращений к базе
	dbPostgres.Close()
	dbMySql.Close()
}
