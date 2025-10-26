package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// generates a bcrypt hash from a plaintext password
// check bcrypt costs for more but 8 is fast/weak, 12-14 strong/slow. 10 defaultcost
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// compares a plaintext password with a stored bcrypt hash
// true if match
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
