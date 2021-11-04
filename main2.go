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

		Name     string `json:"id"`
		Tag      string `json:"tag"`
		Count    uint64 `json:"count"`
		Taxonomy string `json:"taxonomy"`
	}
)

type Tags struct {
	Tag_id           uint `json:"id"`
	Post_id          uint `json:"post_id"`
	Term_taxonomy_id uint
	Tags_prev        []uint `json:"tags_prev"`
	Tags_prev2       uint
	Name             string
	Slug             string
	Count            uint
	Taxonomy         string

	New_id uint
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	oldDatabaseHost := os.Getenv("OLD_DATABASE_HOST")
	oldDatabaseUser := os.Getenv("OLD_DATABASE_USER")
	oldDatabasePass := os.Getenv("OLD_DATABASE_PASS")
	oldDatabaseBase := os.Getenv("OLD_DATABASE_BASE")

	newDatabaseDSN := os.Getenv("NEW_DATABASE_DSN")

	maria, err := client.Connect(oldDatabaseHost, oldDatabaseUser, oldDatabasePass, oldDatabaseBase)
	if err != nil {
		log.Fatal(err)
	}

	defer maria.Close()

	postgres, err := pgxpool.Connect(context.Background(), newDatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	defer postgres.Close()

	result, err := maria.Execute("select id, post_author, post_title, post_date, post_content from wp_posts where post_status = 'publish' and post_type = 'post' order by id;")
	if err != nil {
		log.Fatal(err)
	}

	for i := range result.Values {
		post := Post{}

		post.OldID, err = result.GetUintByName(i, "id")
		if err != nil {
			log.Fatal(err)
		}
		post.AuthorID, err = result.GetUintByName(i, "post_author")
		if err != nil {
			log.Fatal(err)
		}
		post.Title, err = result.GetStringByName(i, "post_title")
		if err != nil {
			log.Fatal(err)
		}

		date, err := result.GetStringByName(i, "post_date")
		if err != nil {
			log.Fatal(err)
		}
		post.Date, err = time.Parse("2006-01-02 15:04:05", date)
		if err != nil {
			log.Fatal(err)
		}

		post.Content, err = result.GetStringByName(i, "post_content")
		if err != nil {
			log.Fatal(err)
		}

		metaImage, err := maria.Execute("select wp_posts.guid as image from wp_postmeta left join wp_posts on wp_postmeta.meta_value = wp_posts.id where post_id = ? and meta_key = '_thumbnail_id';", post.OldID)
		if err != nil {
			log.Fatal(err)
		}

		if len(metaImage.Values) < 1 {
			log.Fatalf("no image for post %d\n", post.OldID)
		}

		for i := range metaImage.Values {
			post.Image, err = metaImage.GetStringByName(i, "image")
			if err != nil {
				log.Fatal(err)
			}
		}

		metaShort, err := maria.Execute("select meta_value as description from wp_postmeta where meta_key = '_yoast_wpseo_metadesc' and post_id = ?;", post.OldID)
		if err != nil {
			log.Fatal(err)
		}

		if len(metaImage.Values) < 1 {
			log.Fatalf("no short description for post %d\n", post.OldID)
		}

		for i := range metaShort.Values {
			post.Description, err = metaShort.GetStringByName(i, "description")
			if err != nil {
				log.Fatal(err)
			}
		}

		_, err = postgres.Exec(context.Background(), "insert into posts (old_id, author_id, date, content, description, title, image) Values ($1, $2, $3, $4, $5, $6, $7)", post.OldID, post.AuthorID, post.Date, post.Content, post.Description, post.Title, post.Image)
		if err != nil {
			log.Fatal(err)
		}

		metaImage.Close()
		metaShort.Close()
	}

	result.Close()
}
