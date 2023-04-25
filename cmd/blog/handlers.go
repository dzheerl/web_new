package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type indexPage struct {
	FeaturedPost   []*featuredPostData
	MostRecentPost []*mostRecentPostData
}

type postPage struct {
	Title       string `db:"title"`
	Subtitle    string `db:"subtitle"`
	Imgphoto    string `db:"image_url"`
	Description string `db:"content"`
}

type featuredPostData struct {
	PostID       string `db:"post_id"`
	Title        string `db:"title"`
	Subtitle     string `db:"subtitle"`
	Author       string `db:"author"`
	Author_url   string `db:"author_url"`
	Publish_date string `db:"publish_date"`
	Image_url    string `db:"image_url"`
	Theme        string `db:"theme"`
	PostURL      string
}

type mostRecentPostData struct {
	PostID       string `db:"post_id"`
	Title        string `db:"title"`
	Subtitle     string `db:"subtitle"`
	Author       string `db:"author"`
	Author_url   string `db:"author_url"`
	Publish_date string `db:"publish_date"`
	Image_url    string `db:"image_url"`
	PostURL      string
}

func index(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		featuredPosts, err := featuredPost(db)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		mostRecentPosts, err := mostRecentPosts(db) // вызов функции mostRecentPosts

		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}

		ts, err := template.ParseFiles("pages/index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		data := indexPage{
			FeaturedPost:   featuredPosts,   // использование featuredPosts
			MostRecentPost: mostRecentPosts, // использование mostRecentPosts
		}

		err = ts.Execute(w, data)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err.Error())
			return
		}

		log.Println("Request completed successfully")
	}
}

func post(db *sqlx.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		postIDStr := mux.Vars(r)["postID"] // Получаем postID в виде строки из параметров урла

		postID, err := strconv.Atoi(postIDStr) // Конвертируем строку postID в число
		if err != nil {
			http.Error(w, "Invalid post id", 403)
			log.Println(err)
			return
		}

		post, err := postByID(db, postID)
		if err != nil {
			if err == sql.ErrNoRows {
				// sql.ErrNoRows возвращается, когда в запросе к базе не было ничего найдено
				// В таком случае мы возвращем 404 (not found) и пишем в тело, что ордер не найден
				http.Error(w, "Post not found", 404)
				log.Println(err)
				return
			}

			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		ts, err := template.ParseFiles("pages/post.html")
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		err = ts.Execute(w, post)
		if err != nil {
			http.Error(w, "Internal Server Error", 500)
			log.Println(err)
			return
		}

		log.Println("Request completed successfully")
	}
}

func featuredPost(db *sqlx.DB) ([]*featuredPostData, error) {
	const query = `
		SELECT
		  post_id,
		  title,
			subtitle,
			author,
			author_url,
			publish_date,
			image_url,
			COALESCE(theme, '') AS theme
		FROM
		  post
		WHERE featured = 1		
	` // Составляем SQL-запрос для получения записей для секции featured-posts

	var posts []*featuredPostData

	err := db.Select(&posts, query) // Делаем запрос в базу данных
	if err != nil {                 // Проверяем, что запрос в базу данных не завершился с ошибкой
		return nil, err
	}

	for _, post := range posts {
		post.PostURL = "/post/" + post.PostID // Формируем исходя из ID поста в базе
	}

	fmt.Println(posts)

	return posts, nil
}

func mostRecentPosts(db *sqlx.DB) ([]*mostRecentPostData, error) {
	const query = `
		SELECT
		  post_id,
		  title,
		  subtitle,
		  author,
		  author_url,
		  publish_date,
		  image_url
		FROM
		  post
		WHERE featured = 0		
	` // Составляем SQL-запрос для получения записей для секции featured-posts

	var posts []*mostRecentPostData

	err := db.Select(&posts, query) // Делаем запрос в базу данных
	if err != nil {                 // Проверяем, что запрос в базу данных не завершился с ошибкой
		return nil, err
	}

	for _, post := range posts {
		post.PostURL = "/post/" + post.PostID // Формируем исходя из ID поста в базе
	}

	fmt.Println(posts)

	return posts, nil
}

func postByID(db *sqlx.DB, postID int) (postPage, error) {
	const query = `
		SELECT
			title,
			subtitle,
			image_url,
			content
		FROM
			` + "`post`" + `
		WHERE
			post_id = ?
	`
	// В SQL-запросе добавились параметры, как в шаблоне. ? означает параметр, который мы передаем в запрос ниже

	var post postPage

	// Обязательно нужно передать в параметрах postID
	err := db.Get(&post, query, postID)
	if err != nil {
		return postPage{}, err
	}

	return post, nil
}
