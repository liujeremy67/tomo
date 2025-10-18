package models

import "time"

// Session represents a single Pomodoro or focus session
type Session struct {
	ID       uint       `json:"id" db:"id"`
	UserID   uint       `json:"user_id" db:"user_id"`
	StartAt  time.Time  `json:"start_at" db:"start_at"`
	EndAt    *time.Time `json:"end_at,omitempty" db:"end_at"` // nullable
	Duration *int       `json:"duration,omitempty" db:"duration"`
}
