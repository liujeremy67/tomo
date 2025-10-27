package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	"tomo/backend/utils"
)

func TestJWTCreateAndValidate(t *testing.T) {
	// Load .env.test if present
	LoadTestEnv()

	fmt.Println("JWT_SECRET:", os.Getenv("JWT_SECRET")) // sanity check

	userID := 1
	email := "test@example.com"
	ttl := time.Hour

	token, err := utils.CreateToken(userID, email, ttl)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	claims, err := utils.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if int((*claims)["sub"].(float64)) != userID {
		t.Errorf("expected user ID %d, got %v", userID, (*claims)["sub"])
	}

	if (*claims)["email"] != email {
		t.Errorf("expected email %s, got %v", email, (*claims)["email"])
	}

	exp := int64((*claims)["exp"].(float64))
	if time.Until(time.Unix(exp, 0)) <= 0 {
		t.Error("token expired too early")
	}
}
