package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type ResultData struct {
	OriginalURL  string
	ShortenedURL string
}

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/url_shortener")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		functions.handleForm(w, r, db)
	})

	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		functions.handleShorten(w, r, db)
	})

	http.HandleFunc("/short/", func(w http.ResponseWriter, r *http.Request) {
		functions.handleRedirect(w, r, db)
	})

	fmt.Println("URL Shortener is running on :3030")
	if err := http.ListenAndServe(":3030", nil); err != nil {
		log.Fatal("Failed to start server", err)
	}
}
