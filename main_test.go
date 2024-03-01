package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"urlshortner/functions"

	"github.com/stretchr/testify/assert"
)

func TestHandleShorten(t *testing.T) {
	// Créer une instance de la base de données pour les tests
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/url_shortener")
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	defer db.Close()

	externalURL := "https://www.facebook.com/?locale=fr_FR"

	// Créer une requête HTTP simulée avec l'URL externe à raccourcir
	req, err := http.NewRequest("POST", "/shorten", strings.NewReader("url="+externalURL))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	rr := httptest.NewRecorder()
	functions.HandleShorten(rr, req, db)
	assert.Equal(t, http.StatusOK, rr.Code, "status code should be 200 OK")

	// Vérifier que la réponse contient le texte attendu
	expectedBody := "URL Shortener Result"
	assert.Contains(t, rr.Body.String(), expectedBody, "response body should contain expected text")
}

func TestHandleRedirect(t *testing.T) {
	// Créer une instance de la base de données pour les tests
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/url_shortener")
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Créer une requête HTTP simulée avec une clé courte valide
	req, err := http.NewRequest("GET", "/short/4WNXVE", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()

	functions.HandleRedirect(rr, req, db)
	assert.Equal(t, http.StatusMovedPermanently, rr.Code, "status code should be 301 Moved Permanently")

	actualRedirectURL := rr.Header().Get("Location")
	expectedRedirectURL := "https://dev.to/envitab/how-to-build-a-url-shortener-with-go-5hn5"
	assert.Equal(t, expectedRedirectURL, actualRedirectURL, "redirect location should match expected URL")
}

func TestHandleRedirectIncrementClickCount(t *testing.T) {
	// Créer une instance de la base de données pour les tests
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/url_shortener")
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Créer une requête HTTP simulée avec une clé courte valide
	shortKey := "pAK89O"
	req, err := http.NewRequest("GET", "/short/"+shortKey, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Enregistrer le compteur de clics avant l'appel à la redirection
	var initialClickCount int
	err = db.QueryRow("SELECT get_clicked FROM urls WHERE short_key = ?", shortKey).Scan(&initialClickCount)
	if err != nil {
		t.Fatalf("failed to retrieve initial click count: %v", err)
	}

	rr := httptest.NewRecorder()
	functions.HandleRedirect(rr, req, db)
	assert.Equal(t, http.StatusMovedPermanently, rr.Code, "status code should be 301 Moved Permanently")

	// Vérifier que le compteur de clics a été incrémenté dans la base de données
	var updatedClickCount int
	err = db.QueryRow("SELECT get_clicked FROM urls WHERE short_key = ?", shortKey).Scan(&updatedClickCount)
	if err != nil {
		t.Fatalf("failed to retrieve updated click count: %v", err)
	}

	assert.Equal(t, initialClickCount+1, updatedClickCount, "click count should be incremented after redirection")
}
