package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"urlshortner/functions"

	"github.com/stretchr/testify/assert"
)

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
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/url_shortener")
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	defer db.Close()

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

func TestGenerateShortKey(t *testing.T) {
	generatedKey := functions.GenerateShortKey()
	//var generatedKey = "JknJwn"
	expectedLength := 6
	assert.Equal(t, expectedLength, len(generatedKey), "generated short key should have expected length")

	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/url_shortener")
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	var existingKey string
	err = db.QueryRow("SELECT short_key FROM urls WHERE short_key = ?", generatedKey).Scan(&existingKey)
	switch {
	case err == sql.ErrNoRows:
	// La clé n'existe pas dans la base de données, c'est bon
	case err != nil:
		t.Fatalf("error checking for existing key: %v", err)
	default:
		t.Fatalf("generated key already exists in database: %s", existingKey)
	}
}
