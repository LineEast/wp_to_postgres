package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
	"github.com/valyala/bytebufferpool"
)

type (
	Post struct {
		ID 			 uint 	`json:"id"`
		OldID        uint   `json:"oldId"`
		Author       uint   `json:"author"`
		Date         string `json:"date"`
		Content      string `json:"content"`
		ContentShort string `json:"ContentShort"`
		Title        string `json:"title"`
		Image        string `json:"image"`
		TagsID       []uint `json:"tagsID"`

		Views uint `json:"PostViewsCount"`
	}

	Tag struct {
		ID    		uint64 	`json:"id"`
		Name  		string 	`json:"name"`
		Alias 		string 	`json:"alias"`
		Type  		string 	`json:"type"`
	}
)

var (
	re = regexp.MustCompile("(<!--)(.{0,2})(wp:)(.{4,32})(-->)")
	reg = regexp.MustCompile("(\n)")
)

func startBase() (dbMySql *sql.DB, dbPostgres *pgxpool.Pool) {
	err := godotenv.Load()
	errorShort(err)

	oldDatabaseDSN := os.Getenv("OLD_DB")
	newDatabaseDSN := os.Getenv("PSQL")
	dbMySql, err = sql.Open("mysql", oldDatabaseDSN)
	errorShort(err)
	dbPostgres, err = pgxpool.Connect(context.Background(), newDatabaseDSN)
	errorShort(err)
	return
}

func delComments(post string) (content string) {
	bb := bytebufferpool.Get()

	indexes := re.FindAllStringIndex(post, -1)

	start, end := 0, 0

	for i := range indexes {
		start = indexes[i][0]
		bb.WriteString(post[end:start])
		end = indexes[i][1]
	}

	content = bb.String()
	bytebufferpool.Put(bb)

	return
}

func delStrings(post string) (content string) {
	bb := bytebufferpool.Get()

	indexes := reg.FindAllStringIndex(post, -1)

	start, end := 0, 0

	for i := range indexes {
		start = indexes[i][0]
		bb.WriteString(post[end:start])
		end = indexes[i][1]
	}

	content = bb.String()
	bytebufferpool.Put(bb)

	return
}

func oneStringSolutionByCmetallo(s string) (res string) {
	res = re.ReplaceAllString(s, "")
	res = reg.ReplaceAllString(res, "")
	return
}

func allInfo(dbMySql *sql.DB, dbPostgres *pgxpool.Pool) {
	posts, err := dbMySql.Query("select id, post_author, post_title, post_date, post_content, post_excerpt from wp_posts where post_status = 'publish' and post_type = 'post' order by id;")
	errorShort(err)

	for posts.Next() {
		post := new(Post)
		err = posts.Scan(&post.OldID, &post.Author, &post.Title, &post.Date, &post.Content, &post.ContentShort)
		errorShort(err)
		post.Content = oneStringSolutionByCmetallo(post.Content)

		if post.ContentShort == "" {
			metaValue, err := dbMySql.Query("select meta_value as description from wp_postmeta where meta_key = '_yoast_wpseo_metadesc' and post_id = ?;", post.OldID)
			errorShort(err)
			if metaValue.Next() {
				err = metaValue.Scan(&post.ContentShort)
				errorShort(err)
			}
			metaValue.Close()
		}

		metaImage, err := dbMySql.Query("select wp_posts.guid as image from wp_postmeta left join wp_posts on wp_postmeta.meta_value = wp_posts.id where post_id = ? and meta_key = '_thumbnail_id';", post.OldID)
		errorShort(err)
		if metaImage.Next() {
			err = metaImage.Scan(&post.Image)
			errorShort(err)
		}
		metaImage.Close()

		views, err := dbMySql.Query("select meta_value from wp_postmeta where meta_key = 'post_views_count' and post_id = ?", post.ID)
		if views.Next() {
			views.Scan(&post.Views)
		}

		err = dbPostgres.QueryRow(context.Background(), "insert into posts (old_id, author_id, date, content, description, title, image, views) values ($1, $2, $3, $4, $5, $6, $7, $8) returning id", post.OldID, post.Author, post.Date, post.Content, post.ContentShort, post.Title, post.Image, post.Views).Scan(&post.ID)
		errorShort(err)
		
		tagsTerm, err := dbMySql.Query("select name, slug, taxonomy from wp_terms left join wp_term_taxonomy on wp_terms.term_id = wp_term_taxonomy.term_id left join wp_term_relationships on wp_term_taxonomy.term_taxonomy_id = wp_term_relationships.term_taxonomy_id where object_id = ? and (taxonomy = 'category' or taxonomy = 'post_tag')", post.OldID)
		errorShort(err)
		for tagsTerm.Next() {
			tag := Tag{}
			err = tagsTerm.Scan(&tag.Name, &tag.Alias, &tag.Type)
			errorShort(err)

			count, err := dbPostgres.Query(context.Background(), "select id from tags where name = $1 and alias = $2;", tag.Name, tag.Alias)
			errorShort(err)

			if count.Next() {
				err = count.Scan(&tag.ID)
				errorShort(err)
			} else {
				err = dbPostgres.QueryRow(context.Background(), "insert into tags (name, alias, type) values ($1, $2, $3) returning id;", tag.Name, tag.Alias, tag.Type).Scan(&tag.ID)
				errorShort(err)
			}
			count.Close()

			_, err = dbPostgres.Exec(context.Background(), "insert into posts_tags (post_id, tag_id) values ($1, $2);", post.ID, tag.ID)
			errorShort(err)
		}
		views.Close()
		tagsTerm.Close()
		
		fmt.Println(post.ID, post.OldID)
	}
	posts.Close()
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
	//Закрытие обращений к базе
	dbPostgres.Close()
	dbMySql.Close()
}