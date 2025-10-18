// User struct definition

package models

// User represents a single user of the app
type User struct {
	ID     uint   `json:"id" db:"id"`
	Handle string `json:"handle" db:"handle"`
	Name   string `json:"name" db:"name"`
}
