package models

import (
	"database/sql"
	"time"
)

// User represents a user in the database
type User struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	Username    string    `json:"username,omitempty"`
	GoogleID    string    `json:"google_id"`
	DisplayName string    `json:"display_name,omitempty"`
	PictureURL  string    `json:"picture_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// CREATE: inserts a new blank user with only Google ID and email
func CreateUser(db *sql.DB, googleID, email string) (User, error) {
	var user User

	err := db.QueryRow(
		`INSERT INTO users (email, google_id, created_at)
		 VALUES ($1, $2, NOW())
		 RETURNING id, email, username, google_id, display_name, picture_url, created_at`,
		email, googleID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.GoogleID, &user.DisplayName, &user.PictureURL, &user.CreatedAt)
	return user, err
}

// READ: fetch a user by email
func GetUserByEmail(db *sql.DB, email string) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, email, username, google_id, display_name, picture_url, created_at
		 FROM users
		 WHERE email=$1`,
		email,
	).Scan(&user.ID, &user.Email, &user.Username, &user.GoogleID, &user.DisplayName, &user.PictureURL, &user.CreatedAt)
	return user, err
}

// READ: fetch a user by Google ID
func GetUserByGoogleID(db *sql.DB, googleID string) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, email, username, google_id, display_name, picture_url, created_at
		 FROM users
		 WHERE google_id=$1`,
		googleID,
	).Scan(&user.ID, &user.Email, &user.Username, &user.GoogleID, &user.DisplayName, &user.PictureURL, &user.CreatedAt)
	return user, err
}

// READ: fetch a user by ID
func GetUserByID(db *sql.DB, id int) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, email, username, google_id, display_name, picture_url, created_at
		 FROM users
		 WHERE id=$1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Username, &user.GoogleID, &user.DisplayName, &user.PictureURL, &user.CreatedAt)
	return user, err
}

// READ: fetch a user by username
func GetUserByUsername(db *sql.DB, username string) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, email, username, google_id, display_name, picture_url, created_at
		 FROM users
		 WHERE username=$1`,
		username,
	).Scan(&user.ID, &user.Email, &user.Username, &user.GoogleID, &user.DisplayName, &user.PictureURL, &user.CreatedAt)
	return user, err
}

// READ: check if a username already exists
func UsernameExists(db *sql.DB, username string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)`,
		username,
	).Scan(&exists)
	return exists, err
}

// UPDATE: update user's profile (username, display_name, picture_url)
func UpdateProfile(db *sql.DB, id int, username, displayName, pictureURL string) error {
	_, err := db.Exec(
		`UPDATE users
		 SET username=$1, display_name=$2, picture_url=$3
		 WHERE id=$4`,
		username, displayName, pictureURL, id,
	)
	return err
}

// UPDATE: update user's email
func UpdateUserEmail(db *sql.DB, id int, newEmail string) error {
	_, err := db.Exec(
		`UPDATE users
		 SET email=$1
		 WHERE id=$2`,
		newEmail, id,
	)
	return err
}

// UPDATE: update user's username
func UpdateUsername(db *sql.DB, id int, newUsername string) error {
	_, err := db.Exec(
		`UPDATE users
		 SET username=$1
		 WHERE id=$2`,
		newUsername, id,
	)
	return err
}

// UPDATE: update user's profile picture
func UpdatePictureURL(db *sql.DB, id int, pictureURL string) error {
	_, err := db.Exec(
		`UPDATE users
		 SET picture_url=$1
		 WHERE id=$2`,
		pictureURL, id,
	)
	return err
}

// UPDATE: update user's display name
func UpdateDisplayName(db *sql.DB, id int, displayName string) error {
	_, err := db.Exec(
		`UPDATE users
		 SET display_name=$1
		 WHERE id=$2`,
		displayName, id,
	)
	return err
}

// DELETE: remove user by ID
func DeleteUser(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}
