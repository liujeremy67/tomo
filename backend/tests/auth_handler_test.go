package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"login-auth-template/handlers"
)

func TestRegisterAndLogin(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	h := &handlers.AuthHandler{DB: db}

	// --- Register ---
	registerBody := map[string]string{
		"email":    "test@example.com",
		"username": "tester",
		"password": "secret123",
	}
	bodyBytes, _ := json.Marshal(registerBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", res.StatusCode)
	}

	// --- Login ---
	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "secret123",
	}
	bodyBytes, _ = json.Marshal(loginBody)

	req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	h.Login(w, req)

	res = w.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", res.StatusCode)
	}

	// Parse response JSON
	var resp map[string]string
	json.NewDecoder(res.Body).Decode(&resp)
	token := resp["token"]

	if token == "" {
		t.Fatalf("expected token in login response")
	}
}
