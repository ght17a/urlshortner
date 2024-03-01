package main_test

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleShorten(t *testing.T) {
	// Créer une instance de la base de données pour les tests
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/url_shortener_test")
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Créer une requête HTTP simulée
	req, err := http.NewRequest("POST", "/shorten", strings.NewReader("url=http://example.com"))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Créer un enregistreur de réponse HTTP simulé
	rr := httptest.NewRecorder()

	// Appeler la fonction handleShorten avec la requête et l'enregistreur simulés
	handleShorten(rr, req, db)

	// Vérifier que la réponse HTTP est celle attendue
	assert.Equal(t, http.StatusOK, rr.Code, "status code should be 200 OK")

	// Vérifier que la réponse contient le texte attendu
	expectedBody := "Shortened URL:"
	assert.Contains(t, rr.Body.String(), expectedBody, "response body should contain expected text")
}

func TestHandleRedirect(t *testing.T) {
	// Créer une instance de la base de données pour les tests
	db, err := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/url_shortener_test")
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	defer db.Close()

	// Créer une requête HTTP simulée avec une clé courte valide
	req, err := http.NewRequest("GET", "/short/abcd123", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Créer un enregistreur de réponse HTTP simulé
	rr := httptest.NewRecorder()

	// Appeler la fonction handleRedirect avec la requête et l'enregistreur simulés
	handleRedirect(rr, req, db)

	// Vérifier que la réponse HTTP est une redirection permanente (status code 301)
	assert.Equal(t, http.StatusMovedPermanently, rr.Code, "status code should be 301 Moved Permanently")

	// Vérifier que la réponse contient le header de redirection
	expectedHeader := fmt.Sprintf("Location: %s", "http://example.com")
	assert.Contains(t, rr.Header().Get("Location"), expectedHeader, "redirect location should match expected URL")
}
