//go:build !old

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/go-mysql-org/go-mysql/client"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/joho/godotenv"
)

type (
	Post struct {
		ID    uint64 `json:"id"`
		OldID uint64 `json:"oldId"`

		AuthorID uint64 `json:"authorId"`

		Title   string `json:"title"`
		Image   string `json:"img"`
		Content string `json:"content"`

		Date        time.Time `json:"date"`
		Description string    `json:"description"`
	}

	Tag struct {
		ID uint64 `json:"id"`

		Name  string `json:"name"`
		Alias string `json:"alias"`
		Type  string `json:"type"`
	}
)


func main() {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	oldDatabaseHost := os.Getenv("OLD_DATABASE_HOST")
	oldDatabaseUser := os.Getenv("OLD_DATABASE_USER")
	oldDatabasePass := os.Getenv("OLD_DATABASE_PASS")
	oldDatabaseBase := os.Getenv("OLD_DATABASE_BASE")

	newDatabaseDSN := os.Getenv("NEW_DATABASE_DSN")

	maria, err := client.Connect(oldDatabaseHost, oldDatabaseUser, oldDatabasePass, oldDatabaseBase)
	if err != nil {
		panic(err)
	}
	defer maria.Close()

	postgres, err := pgxpool.Connect(context.Background(), newDatabaseDSN)
	if err != nil {
		panic(err)
	}
	defer postgres.Close()

	posts, err := maria.Execute("select id, post_author, post_title, post_date, post_content, post_excerpt from wp_posts where post_status = 'publish' and post_type = 'post' order by id;")
	if err != nil {
		panic(err)
	}

	for i := range posts.Values {
		post := Post{}

		post.OldID, err = posts.GetUintByName(i, "id")
		if err != nil {
			panic(err)
		}
		post.AuthorID, err = posts.GetUintByName(i, "post_author")
		if err != nil {
			panic(err)
		}
		post.Title, err = posts.GetStringByName(i, "post_title")
		if err != nil {
			panic(err)
		}

		date, err := posts.GetStringByName(i, "post_date")
		if err != nil {
			panic(err)
		}
		post.Date, err = time.Parse("2006-01-02 15:04:05", date)
		if err != nil {
			panic(err)
		}

		post.Content, err = posts.GetStringByName(i, "post_content")
		if err != nil {
			panic(err)
		}

		post.Description, err = posts.GetStringByName(i, "post_excerpt")
		if err != nil {
			panic(err)
		}

		if post.Description == "" {
			metaShort, err := maria.Execute("select meta_value as description from wp_postmeta where meta_key = '_yoast_wpseo_metadesc' and post_id = ?;", post.OldID)
			if err != nil {
				panic(err)
			}

			if len(metaShort.Values) < 1 {
				log.Fatalf("no short description for post %d\n", post.OldID)
			}

			for i := range metaShort.Values {
				post.Description, err = metaShort.GetStringByName(i, "description")
				if err != nil {
					panic(err)
				}
			}
			metaShort.Close()
		}

		metaImage, err := maria.Execute("select wp_posts.guid as image from wp_postmeta left join wp_posts on wp_postmeta.meta_value = wp_posts.id where post_id = ? and meta_key = '_thumbnail_id';", post.OldID)
		if err != nil {
			panic(err)
		}

		if len(metaImage.Values) < 1 {
			log.Fatalf("no image for post %d\n", post.OldID)
		}

		for i := range metaImage.Values {
			post.Image, err = metaImage.GetStringByName(i, "image")
			if err != nil {
				panic(err)
			}
		}
		metaImage.Close()

		err = postgres.QueryRow(context.Background(), "insert into posts (old_id, author_id, date, content, description, title, image) values ($1, $2, $3, $4, $5, $6, $7) returning id", post.OldID, post.AuthorID, post.Date, post.Content, post.Description, post.Title, post.Image).Scan(&post.ID)
		if err != nil {
			panic(err)
		}

		tagsTerm, err := maria.Execute("select name, slug, taxonomy from wp_terms left join wp_term_taxonomy on wp_terms.term_id = wp_term_taxonomy.term_id left join wp_term_relationships on wp_term_taxonomy.term_taxonomy_id = wp_term_relationships.term_taxonomy_id where object_id = ? and (taxonomy = 'category' or taxonomy = 'post_tag')", post.OldID)
		if err != nil {
			panic(err)
		}

		for i := range tagsTerm.Values {
			tag := Tag{}

			tag.Name, err = tagsTerm.GetStringByName(i, "name")
			if err != nil {
				panic(err)
			}
			tag.Alias, err = tagsTerm.GetStringByName(i, "slug")
			if err != nil {
				panic(err)
			}
			tag.Type, err = tagsTerm.GetStringByName(i, "taxonomy")
			if err != nil {
				panic(err)
			}

			rows, err := postgres.Query(context.Background(), "select id from tags where name = $1 and alias = $2;", tag.Name, tag.Alias)
			if err != nil {
				panic(err)
			}

			if rows.Next() {
				rows.Scan(&tag.ID)
			} else {
				err = postgres.QueryRow(context.Background(), "insert into tags (name, alias, type) values ($1, $2, $3) returning id;", tag.Name, tag.Alias, tag.Type).Scan(&tag.ID)
				if err != nil {
					panic(err)
				}
			}

			rows.Close()

			_, err = postgres.Exec(context.Background(), "insert into posts_tags (post_id, tag_id) values ($1, $2);", post.ID, tag.ID)
			if err != nil {
				panic(err)
			}
		}

		tagsTerm.Close()
	}

	posts.Close()
}
