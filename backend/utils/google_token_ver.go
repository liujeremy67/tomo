package utils

import (
	"context"
	"errors"
	"os"

	"google.golang.org/api/idtoken"
)

// GoogleTokenPayload contains the verified user information from Google
type GoogleTokenPayload struct {
	GoogleID      string
	Email         string
	EmailVerified bool
}

// VerifyGoogleToken validates a Google ID token and extracts user info
func VerifyGoogleToken(ctx context.Context, idToken string) (*GoogleTokenPayload, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return nil, errors.New("GOOGLE_CLIENT_ID not set")
	}

	// Verify the token with Google
	payload, err := idtoken.Validate(ctx, idToken, clientID)
	if err != nil {
		return nil, errors.New("invalid Google token")
	}

	// Extract claims
	email, _ := payload.Claims["email"].(string)
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	googleID := payload.Subject // "sub" claim is the Google user ID

	// Ensure email is verified
	if !emailVerified {
		return nil, errors.New("email not verified by Google")
	}

	if email == "" || googleID == "" {
		return nil, errors.New("missing required claims in token")
	}

	return &GoogleTokenPayload{
		GoogleID:      googleID,
		Email:         email,
		EmailVerified: emailVerified,
	}, nil
}
