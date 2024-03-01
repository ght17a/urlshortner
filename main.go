package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/urlshortner")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleForm(w, r)
	})
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		handleShorten(w, r, db)
	})
	http.HandleFunc("/short/", func(w http.ResponseWriter, r *http.Request) {
		handleRedirect(w, r, db)
	})

	fmt.Println("URL Shortener is running on :3030")
	if err := http.ListenAndServe(":3030", nil); err != nil {
		log.Fatal("Failed to start server", err)
	}
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/shorten", http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
        <!DOCTYPE html>
        <html>
        <head>
            <title>URL Shortener</title>
        </head>
        <body>
            <h2>URL Shortener</h2>
            <form method="post" action="/shorten">
                <input type="url" name="url" placeholder="Enter a URL" required>
                <input type="submit" value="Shorten">
            </form>
        </body>
        </html>
    `)
}

func handleShorten(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	shortKey := generateShortKey()
	shortenedURL := fmt.Sprintf("http://localhost:3030/short/%s", shortKey)

	stmt, err := db.Prepare("INSERT INTO urls(short_key, original_url, shortened_url) VALUES(?, ?, ?)")
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(shortKey, originalURL, shortenedURL)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
        <!DOCTYPE html>
        <html>
        <head>
            <title>URL Shortener</title>
        </head>
        <body>
            <h2>URL Shortener</h2>
            <p>Original URL: `, originalURL, `</p>
            <p>Shortened URL: <a href="`, shortenedURL, `">`, shortenedURL, `</a></p>
        </body>
        </html>
    `)
}

func handleRedirect(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	shortKey := strings.TrimPrefix(r.URL.Path, "/short/")
	if shortKey == "" {
		http.Error(w, "Shortened key is missing", http.StatusBadRequest)
		return
	}

	var originalURL string
	err := db.QueryRow("SELECT original_url FROM urls WHERE short_key = ?", shortKey).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Shortened key not found", http.StatusNotFound)
		} else {
			http.Error(w, "Server error", http.StatusInternalServerError)
			log.Println(err)
		}
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func generateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	rand.Seed(time.Now().UnixNano())
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}
