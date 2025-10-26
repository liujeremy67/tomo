package tests

import (
	"testing"

	"login-auth-template/models"
)

func TestUserCRUD(t *testing.T) {

	// Clean DB before starting
	db := SetupTestDB(t)

	// -------- CREATE --------
	user, err := models.CreateUser(db, "alice@example.com", "alice", "hashedpassword")
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("CreateUser returned invalid ID")
	}
	if user.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %s", user.Email)
	}

	// -------- READ by email --------
	u, err := models.GetUserByEmail(db, "alice@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail failed: %v", err)
	}
	if u.ID != user.ID {
		t.Fatalf("GetUserByEmail returned wrong user ID: got %d, want %d", u.ID, user.ID)
	}

	// -------- READ by ID --------
	u2, err := models.GetUserByID(db, user.ID)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}
	if u2.Email != user.Email {
		t.Fatalf("GetUserByID returned wrong email: got %s, want %s", u2.Email, user.Email)
	}

	// -------- UPDATE email --------
	newEmail := "alice2@example.com"
	if err := models.UpdateUserEmail(db, user.ID, newEmail); err != nil {
		t.Fatalf("UpdateUserEmail failed: %v", err)
	}
	u3, _ := models.GetUserByID(db, user.ID)
	if u3.Email != newEmail {
		t.Fatalf("Email not updated: got %s, want %s", u3.Email, newEmail)
	}

	// -------- UPDATE username --------
	newUsername := "alice2"
	if err := models.UpdateUsername(db, user.ID, newUsername); err != nil {
		t.Fatalf("UpdateUsername failed: %v", err)
	}
	u4, _ := models.GetUserByID(db, user.ID)
	if u4.Username != newUsername {
		t.Fatalf("Username not updated: got %s, want %s", u4.Username, newUsername)
	}

	// -------- UPDATE password --------
	newHash := "newhashedpassword"
	if err := models.UpdatePassword(db, user.ID, newHash); err != nil {
		t.Fatalf("UpdatePassword failed: %v", err)
	}
	u5, _ := models.GetUserByID(db, user.ID)
	if u5.PasswordHash != newHash {
		t.Fatalf("PasswordHash not updated: got %s, want %s", u5.PasswordHash, newHash)
	}

	// -------- DELETE --------
	if err := models.DeleteUser(db, user.ID); err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
	_, err = models.GetUserByID(db, user.ID)
	if err == nil {
		t.Fatal("Deleted user still exists")
	}
}
