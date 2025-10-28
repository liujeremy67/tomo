package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"tomo/backend/middleware"
	"tomo/backend/models"
	"tomo/backend/utils"
)

type SessionHandler struct {
	DB *sql.DB
}

type CreateSessionRequest struct {
	StartTime string `json:"start_time"` // ISO 8601 format
	EndTime   string `json:"end_time"`   // ISO 8601 format
}

// POST /sessions — create a new focus session
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Parse timestamps
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "invalid start_time format (use ISO 8601)", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "invalid end_time format (use ISO 8601)", http.StatusBadRequest)
		return
	}

	// Validation
	if endTime.Before(startTime) {
		http.Error(w, "end_time must be after start_time", http.StatusBadRequest)
		return
	}

	duration := endTime.Sub(startTime).Minutes()
	if duration < 1 {
		http.Error(w, "session must be at least 1 minute long", http.StatusBadRequest)
		return
	}
	if duration > 1440 { // 24 hours
		http.Error(w, "session cannot exceed 24 hours", http.StatusBadRequest)
		return
	}

	// Create session
	session, err := models.CreateSession(h.DB, user.UserID, startTime, endTime)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, session)
}

// GET /sessions — get all sessions for the authenticated user
func (h *SessionHandler) GetMySessions(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Pagination (default: 50 sessions per page)
	limit := 50
	offset := 0

	sessions, err := models.GetSessionsByUserID(h.DB, user.UserID, limit, offset)
	if err != nil {
		http.Error(w, "failed to fetch sessions", http.StatusInternalServerError)
		return
	}

	// Get stats
	stats, err := models.GetUserSessionStats(h.DB, user.UserID)
	if err != nil {
		http.Error(w, "failed to fetch stats", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"sessions": sessions,
		"stats":    stats,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// GET /sessions/{id} — get a specific session by ID
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := r.PathValue("id")
	if sessionID == "" {
		http.Error(w, "session id is required", http.StatusBadRequest)
		return
	}

	var id int
	if _, err := fmt.Sscanf(sessionID, "%d", &id); err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	session, err := models.GetSessionByID(h.DB, id)
	if err == sql.ErrNoRows {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Verify ownership
	if session.UserID != user.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	utils.WriteJSON(w, http.StatusOK, session)
}

// DELETE /sessions/{id} — delete a session
func (h *SessionHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	sessionID := r.PathValue("id")
	var id int
	if _, err := fmt.Sscanf(sessionID, "%d", &id); err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}

	// Check ownership
	session, err := models.GetSessionByID(h.DB, id)
	if err == sql.ErrNoRows {
		http.Error(w, "session not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if session.UserID != user.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := models.DeleteSession(h.DB, id); err != nil {
		http.Error(w, "failed to delete session", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "session deleted"})
}
