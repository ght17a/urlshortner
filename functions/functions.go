package functions

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type ResultData struct {
	OriginalURL  string
	ShortenedURL string
}

func HandleForm(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Récupérer le nombre de liens raccourcis depuis la base de données
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM urls").Scan(&count)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		log.Println("Error getting link count:", err)
		return
	}

	// Récupérer toutes les URL raccourcies avec le nombre de clics associés
	rows, err := db.Query("SELECT shortened_url, get_clicked FROM urls")
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		log.Println("Error getting shortened URLs:", err)
		return
	}
	defer rows.Close()

	// Stocker les données des URL raccourcies et le nombre de clics dans une structure
	var shortenedURLs []struct {
		ShortenedURL string
		ClickCount   int
	}
	for rows.Next() {
		var shortURL string
		var clickCount int
		if err := rows.Scan(&shortURL, &clickCount); err != nil {
			log.Println("Error scanning rows:", err)
			continue
		}
		shortenedURLs = append(shortenedURLs, struct {
			ShortenedURL string
			ClickCount   int
		}{ShortenedURL: shortURL, ClickCount: clickCount})
	}
	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
	}

	// Lecture du contenu HTML du fichier form.html
	htmlContent, err := ioutil.ReadFile("templates/form.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error reading HTML file:", err)
		return
	}

	// Passer le nombre de liens raccourcis et les données des URL raccourcies comme données de modèle
	data := struct {
		LinkCount     int
		ShortenedURLs []struct {
			ShortenedURL string
			ClickCount   int
		}
	}{
		LinkCount:     count,
		ShortenedURLs: shortenedURLs,
	}

	w.Header().Set("Content-Type", "text/html")
	t, err := template.New("form").Parse(string(htmlContent))
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error parsing HTML template:", err)
		return
	}
	// Exécuter le modèle avec les données de modèle
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error executing HTML template:", err)
		return
	}
}

func HandleShorten(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	shortKey := GenerateShortKey()
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

	// Structure de données pour les informations de résultat
	result := ResultData{
		OriginalURL:  originalURL,
		ShortenedURL: shortenedURL,
	}

	// Analyse du modèle HTML
	tmpl, err := template.ParseFiles("templates/result.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error parsing HTML template:", err)
		return
	}

	// Exécution du modèle HTML avec les données du résultat
	err = tmpl.Execute(w, result)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println("Error executing HTML template:", err)
		return
	}
}

func HandleRedirect(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	// Récupérer la clé courte à partir de l'URL
	shortKey := strings.TrimPrefix(r.URL.Path, "/short/")
	if shortKey == "" {
		http.Error(w, "Shortened key is missing", http.StatusBadRequest)
		return
	}

	// Récupérer l'URL d'origine à partir de la base de données
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

	// Incrémenter le compteur de clics dans la base de données
	_, err = db.Exec("UPDATE urls SET get_clicked = get_clicked + 1 WHERE id = (SELECT sub.id FROM (SELECT * FROM urls) as sub WHERE sub.short_key = ? LIMIT 1) ", shortKey)
	if err != nil {
		log.Println("Error incrementing click count:", err)
	}
}

func GenerateShortKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 6

	rand.Seed(time.Now().UnixNano())
	shortKey := make([]byte, keyLength)
	for i := range shortKey {
		shortKey[i] = charset[rand.Intn(len(charset))]
	}
	return string(shortKey)
}
func HandleRegister(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == "GET" {
		fmt.Fprintf(w, `<h1>Inscription</h1>
                        <form method="post">
                            Username: <input type="text" name="username"/><br>
                            Password: <input type="password" name="password"/><br>
                            <input type="submit" value="Register"/>
                        </form>
                        <p>Vous avez déjà un compte ? <a href="/login">Connexion</a></p>`)
	} else if r.Method == "POST" {
		switch r.Method {
		case "GET":
			fmt.Fprintf(w, "<h1>Inscription</h1><form method='post'>Username: <input type='text' name='username'/><br>Password: <input type='password' name='password'/><br><input type='submit' value='Register'/></form>")
		case "POST":
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Erreur lors du parsing du formulaire", http.StatusInternalServerError)
				return
			}

			username := r.FormValue("username")
			password := r.FormValue("password")

			// Hachage du mot de passe
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Erreur lors du hachage du mot de passe", http.StatusInternalServerError)
				return
			}

			// Insérer l'utilisateur dans la base de données
			_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hashedPassword))
			if err != nil {
				http.Error(w, "Erreur lors de l'inscription de l'utilisateur", http.StatusInternalServerError)
				return
			}

			// Rediriger l'utilisateur vers la page de connexion après l'inscription réussie
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}
func HandleLogin(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	if r.Method == "GET" {
		fmt.Fprintf(w, `<h1>Connexion</h1>
                        <form method="post">
                            Username: <input type="text" name="username"/><br>
                            Password: <input type="password" name="password"/><br>
                            <input type="submit" value="Login"/>
                        </form>
                        <p>Vous n'avez pas de compte ? <a href="/register">Inscription</a></p>`)
	} else if r.Method == "POST" {
		switch r.Method {
		case "GET":
			fmt.Fprintf(w, "<h1>Connexion</h1><form method='post'>Username: <input type='text' name='username'/><br>Password: <input type='password' name='password'/><br><input type='submit' value='Login'/></form>")
		case "POST":
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "Erreur lors du parsing du formulaire", http.StatusInternalServerError)
				return
			}

			username := r.FormValue("username")
			password := r.FormValue("password")

			var dbPassword string
			err = db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&dbPassword)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Utilisateur non trouvé", http.StatusBadRequest)
				} else {
					http.Error(w, "Erreur lors de la recherche de l'utilisateur", http.StatusInternalServerError)
				}
				return
			}

			err = bcrypt.CompareHashAndPassword([]byte(dbPassword), []byte(password))
			if err != nil {
				http.Error(w, "Mot de passe invalide", http.StatusUnauthorized)
				return
			}

			http.Redirect(w, r, "/form", http.StatusSeeOther)
		default:
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		}
	}
}
