package models

import (
	"database/sql"
	"time"
)

// FocusSession represents a completed focus/work session
type FocusSession struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationMinutes int       `json:"duration_minutes"`
	CreatedAt       time.Time `json:"created_at"`
}

// CREATE: insert a new focus session
func CreateSession(db *sql.DB, userID int, startTime, endTime time.Time) (FocusSession, error) {
	var session FocusSession

	// Calculate duration in minutes
	duration := int(endTime.Sub(startTime).Minutes())

	err := db.QueryRow(
		`INSERT INTO focus_sessions (user_id, start_time, end_time, duration_minutes, created_at)
		 VALUES ($1, $2, $3, $4, NOW())
		 RETURNING id, user_id, start_time, end_time, duration_minutes, created_at`,
		userID, startTime, endTime, duration,
	).Scan(&session.ID, &session.UserID, &session.StartTime, &session.EndTime, &session.DurationMinutes, &session.CreatedAt)

	return session, err
}

// READ: get a session by ID
func GetSessionByID(db *sql.DB, sessionID int) (FocusSession, error) {
	var session FocusSession
	err := db.QueryRow(
		`SELECT id, user_id, start_time, end_time, duration_minutes, created_at
		 FROM focus_sessions
		 WHERE id=$1`,
		sessionID,
	).Scan(&session.ID, &session.UserID, &session.StartTime, &session.EndTime, &session.DurationMinutes, &session.CreatedAt)
	return session, err
}

// READ: get all sessions for a user (paginated)
func GetSessionsByUserID(db *sql.DB, userID int, limit, offset int) ([]FocusSession, error) {
	rows, err := db.Query(
		`SELECT id, user_id, start_time, end_time, duration_minutes, created_at
		 FROM focus_sessions
		 WHERE user_id=$1
		 ORDER BY start_time DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []FocusSession
	for rows.Next() {
		var session FocusSession
		if err := rows.Scan(&session.ID, &session.UserID, &session.StartTime, &session.EndTime, &session.DurationMinutes, &session.CreatedAt); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// READ: get session stats for a user (total time, session count)
func GetUserSessionStats(db *sql.DB, userID int) (map[string]interface{}, error) {
	var totalMinutes int
	var sessionCount int

	err := db.QueryRow(
		`SELECT COALESCE(SUM(duration_minutes), 0), COUNT(*)
		 FROM focus_sessions
		 WHERE user_id=$1`,
		userID,
	).Scan(&totalMinutes, &sessionCount)

	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_minutes": totalMinutes,
		"session_count": sessionCount,
		"total_hours":   float64(totalMinutes) / 60.0,
	}, nil
}

// DELETE: remove a session by ID
func DeleteSession(db *sql.DB, sessionID int) error {
	_, err := db.Exec(`DELETE FROM focus_sessions WHERE id=$1`, sessionID)
	return err
}
