package tests

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"login-auth-template/handlers"
	"login-auth-template/models"
	"login-auth-template/utils"
)

func TestUserHandler_GetAndDeleteMe(t *testing.T) {
	db := SetupTestDB(t)
	defer db.Close()

	// Seed a user
	hash, err := utils.HashPassword("password123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user, err := models.CreateUser(db, "unit@example.com", "unituser", hash)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	h := &handlers.UserHandler{DB: db}

	// ---- Test GET /me ----
	req := httptest.NewRequest("GET", "/me", nil)
	ctx := context.WithValue(req.Context(), "user_id", user.ID) // simulate middleware
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.GetMe(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// ---- Test DELETE /me ----
	req2 := httptest.NewRequest("DELETE", "/me", nil)
	ctx2 := context.WithValue(req2.Context(), "user_id", user.ID)
	req2 = req2.WithContext(ctx2)
	w2 := httptest.NewRecorder()

	h.DeleteMe(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d: %s", w2.Code, w2.Body.String())
	}

	// Verify user was deleted
	_, err = models.GetUserByID(db, user.ID)
	if err == nil || err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows after delete, got %v", err)
	}
}
