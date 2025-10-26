package models

import (
	"database/sql"
	"time"
)

// User represents a user in the database
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

// CREATE: inserts a new user into the DB
func CreateUser(db *sql.DB, email, username, passwordHash string) (User, error) {
	var user User
	err := db.QueryRow(
		`INSERT INTO users (email, username, password_hash, created_at)
		 VALUES ($1, $2, $3, NOW())
		 RETURNING id, email, username, password_hash, created_at`,
		email, username, passwordHash,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

// READ: fetch a user by email
func GetUserByEmail(db *sql.DB, email string) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, email, username, password_hash, created_at
		 FROM users
		 WHERE email=$1`,
		email,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
	return user, err
}

// READ: fetch a user by ID
func GetUserByID(db *sql.DB, id int) (User, error) {
	var user User
	err := db.QueryRow(
		`SELECT id, email, username, password_hash, created_at
		 FROM users
		 WHERE id=$1`,
		id,
	).Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash, &user.CreatedAt)
	return user, err
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

// UPDATE: update user's password (expects hashed password)
func UpdatePassword(db *sql.DB, id int, newPasswordHash string) error {
	_, err := db.Exec(
		`UPDATE users
		 SET password_hash=$1
		 WHERE id=$2`,
		newPasswordHash, id,
	)
	return err
}

// DELETE: remove user by ID
func DeleteUser(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM users WHERE id=$1`, id)
	return err
}
