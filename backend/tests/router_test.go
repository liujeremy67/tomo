package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"login-auth-template/models"
	"login-auth-template/routes"
	"login-auth-template/utils"
)

// TestRouterFullPipeline tests the complete JWT auth flow through the router
func TestRouterFullPipeline(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	router := routes.NewRouter(db)

	// --- Test 1: Register a new user ---
	t.Run("Register", func(t *testing.T) {
		body := map[string]string{
			"email":    "router@test.com",
			"username": "routeruser",
			"password": "testpass123",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}

		var user models.User
		if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode user response: %v", err)
		}

		if user.Email != "router@test.com" {
			t.Errorf("expected email router@test.com, got %s", user.Email)
		}
		if user.Username != "routeruser" {
			t.Errorf("expected username routeruser, got %s", user.Username)
		}
	})

	// --- Test 2: Login and get JWT token ---
	var token string
	t.Run("Login", func(t *testing.T) {
		body := map[string]string{
			"email":    "router@test.com",
			"password": "testpass123",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode login response: %v", err)
		}

		token = resp["token"]
		if token == "" {
			t.Fatal("expected token in response")
		}

		// Validate the token structure
		claims, err := utils.ValidateToken(token)
		if err != nil {
			t.Fatalf("token validation failed: %v", err)
		}

		if (*claims)["email"] != "router@test.com" {
			t.Errorf("expected email in claims, got %v", (*claims)["email"])
		}
	})

	// --- Test 3: Access protected route GET /me with valid token ---
	t.Run("GetMe_WithValidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var user models.User
		if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
			t.Fatalf("failed to decode user: %v", err)
		}

		if user.Email != "router@test.com" {
			t.Errorf("expected email router@test.com, got %s", user.Email)
		}
		if user.Username != "routeruser" {
			t.Errorf("expected username routeruser, got %s", user.Username)
		}
	})

	// --- Test 4: Access protected route without token (should fail) ---
	t.Run("GetMe_WithoutToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	// --- Test 5: Access protected route with invalid token ---
	t.Run("GetMe_WithInvalidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		req.Header.Set("Authorization", "Bearer invalid.token.here")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	// --- Test 6: Access protected route with malformed header ---
	t.Run("GetMe_WithMalformedHeader", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		req.Header.Set("Authorization", token) // Missing "Bearer " prefix
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	// --- Test 7: Delete user with valid token ---
	t.Run("DeleteMe_WithValidToken", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/me", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp map[string]string
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp["message"] != "user deleted" {
			t.Errorf("expected 'user deleted', got %s", resp["message"])
		}

		// Verify user is actually deleted from database
		_, err := models.GetUserByEmail(db, "router@test.com")
		if err == nil {
			t.Error("user should be deleted but still exists")
		}
	})

	// --- Test 8: Try to access /me after deletion (token still valid but user gone) ---
	t.Run("GetMe_AfterDeletion", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/me", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should fail because user no longer exists
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})
}

// TestRouterExpiredToken tests behavior with expired JWT
func TestRouterExpiredToken(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	router := routes.NewRouter(db)

	// Create a user first
	hash, _ := utils.HashPassword("testpass")
	user, err := models.CreateUser(db, "expired@test.com", "expireduser", hash)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Create an expired token (negative TTL)
	expiredToken, err := utils.CreateToken(user.ID, user.Email, -1*time.Hour)
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	// Try to access protected route with expired token
	req := httptest.NewRequest("GET", "/me", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", expiredToken))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", w.Code)
	}
}

// TestRouterLoginFailures tests various login failure scenarios
func TestRouterLoginFailures(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	router := routes.NewRouter(db)

	// Create a user
	hash, _ := utils.HashPassword("correctpass")
	_, err := models.CreateUser(db, "fail@test.com", "failuser", hash)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	t.Run("WrongPassword", func(t *testing.T) {
		body := map[string]string{
			"email":    "fail@test.com",
			"password": "wrongpassword",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for wrong password, got %d", w.Code)
		}
	})

	t.Run("NonexistentEmail", func(t *testing.T) {
		body := map[string]string{
			"email":    "nonexistent@test.com",
			"password": "anypassword",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for nonexistent email, got %d", w.Code)
		}
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for invalid JSON, got %d", w.Code)
		}
	})
}

// TestRouterRegisterFailures tests registration failure scenarios
func TestRouterRegisterFailures(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	router := routes.NewRouter(db)

	// Create initial user
	hash, _ := utils.HashPassword("pass")
	_, err := models.CreateUser(db, "existing@test.com", "existinguser", hash)
	if err != nil {
		t.Fatalf("failed to create initial user: %v", err)
	}

	t.Run("DuplicateEmail", func(t *testing.T) {
		body := map[string]string{
			"email":    "existing@test.com",
			"username": "newuser",
			"password": "password123",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for duplicate email, got %d", w.Code)
		}
	})

	t.Run("DuplicateUsername", func(t *testing.T) {
		body := map[string]string{
			"email":    "new@test.com",
			"username": "existinguser",
			"password": "password123",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for duplicate username, got %d", w.Code)
		}
	})
}

// TestRouterConcurrentRequests tests that router handles concurrent auth correctly
func TestRouterConcurrentRequests(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	router := routes.NewRouter(db)

	// Create multiple users
	tokens := make([]string, 3)
	for i := 0; i < 3; i++ {
		email := fmt.Sprintf("concurrent%d@test.com", i)
		username := fmt.Sprintf("user%d", i)

		// Register
		regBody := map[string]string{
			"email":    email,
			"username": username,
			"password": "pass123",
		}
		regBytes, _ := json.Marshal(regBody)
		req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(regBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Login to get token
		loginBody := map[string]string{
			"email":    email,
			"password": "pass123",
		}
		loginBytes, _ := json.Marshal(loginBody)
		req = httptest.NewRequest("POST", "/login", bytes.NewBuffer(loginBytes))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var resp map[string]string
		json.NewDecoder(w.Body).Decode(&resp)
		tokens[i] = resp["token"]
	}

	// Make concurrent requests with different tokens
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func(idx int) {
			req := httptest.NewRequest("GET", "/me", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokens[idx]))
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("user %d: expected 200, got %d", idx, w.Code)
			}

			var user models.User
			json.NewDecoder(w.Body).Decode(&user)
			expectedEmail := fmt.Sprintf("concurrent%d@test.com", idx)
			if user.Email != expectedEmail {
				t.Errorf("user %d: expected email %s, got %s", idx, expectedEmail, user.Email)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}
