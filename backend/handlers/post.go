package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"tomo/backend/middleware"
	"tomo/backend/models"
	"tomo/backend/utils"
)

type PostHandler struct {
	DB *sql.DB
}

type CreatePostRequest struct {
	SessionID  *int     `json:"session_id,omitempty"`
	PostType   string   `json:"post_type"`
	Content    string   `json:"content,omitempty"`
	Title      string   `json:"title,omitempty"`
	MoodRating *int     `json:"mood_rating,omitempty"`
	Visibility string   `json:"visibility"`
	Tags       []string `json:"tags,omitempty"`
}

// POST /posts — create a new reflection post
func (h *PostHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validation
	if req.PostType != "session" && req.PostType != "general" {
		http.Error(w, "post_type must be 'session' or 'general'", http.StatusBadRequest)
		return
	}

	if req.Visibility != "private" && req.Visibility != "public" {
		http.Error(w, "visibility must be 'private' or 'public'", http.StatusBadRequest)
		return
	}

	if req.MoodRating != nil && (*req.MoodRating < 1 || *req.MoodRating > 5) {
		http.Error(w, "mood_rating must be between 1 and 5", http.StatusBadRequest)
		return
	}

	// If post_type is 'session', verify session exists and belongs to user
	if req.PostType == "session" {
		if req.SessionID == nil {
			http.Error(w, "session_id is required for session posts", http.StatusBadRequest)
			return
		}

		session, err := models.GetSessionByID(h.DB, *req.SessionID)
		if err == sql.ErrNoRows {
			http.Error(w, "session not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, "database error", http.StatusInternalServerError)
			return
		}

		if session.UserID != user.UserID {
			http.Error(w, "forbidden: session does not belong to you", http.StatusForbidden)
			return
		}
	}

	// Create post
	post, err := models.CreatePost(h.DB, user.UserID, req.SessionID, req.PostType, req.Content, req.Title, req.MoodRating, req.Visibility)
	if err != nil {
		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	// Add tags if provided
	if len(req.Tags) > 0 {
		if err := models.AddTagsToPost(h.DB, user.UserID, post.ID, req.Tags); err != nil {
			http.Error(w, "failed to add tags", http.StatusInternalServerError)
			return
		}
	}

	// Fetch complete post with tags and media (media will be empty for new posts)
	postDetails, err := models.GetPostWithDetails(h.DB, post.ID)
	if err != nil {
		http.Error(w, "failed to fetch post details", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, postDetails)
}

// GET /posts/user/{id} — get all posts for a specific user
func (h *PostHandler) GetUserPosts(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userIDStr := r.PathValue("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	// Only allow users to view their own posts
	if userID != authUser.UserID {
		http.Error(w, "forbidden: can only view your own posts", http.StatusForbidden)
		return
	}

	// Pagination (default: 20 posts per page)
	limit := 20
	offset := 0

	// Get posts with full details (tags + media)
	posts, err := models.GetPostsWithDetailsByUserID(h.DB, userID, limit, offset)
	if err != nil {
		http.Error(w, "failed to fetch posts", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"posts": posts,
		"count": len(posts),
	})
}

// GET /posts/{id} — get a specific post by ID (with tags and media)
func (h *PostHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	postIDStr := r.PathValue("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	post, err := models.GetPostByID(h.DB, postID)
	if err == sql.ErrNoRows {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Check access: only owner can view private posts
	if post.Visibility == "private" && post.UserID != user.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Get complete post with tags and media
	postDetails, err := models.GetPostWithDetails(h.DB, postID)
	if err != nil {
		http.Error(w, "failed to fetch post details", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, postDetails)
}

// PATCH /posts/{id} — update a post
func (h *PostHandler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	postIDStr := r.PathValue("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	// Check ownership
	post, err := models.GetPostByID(h.DB, postID)
	if err == sql.ErrNoRows {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if post.UserID != user.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Use existing values if not provided
	content := req.Content
	if content == "" {
		content = post.Content
	}
	title := req.Title
	if title == "" {
		title = post.Title
	}
	visibility := req.Visibility
	if visibility == "" {
		visibility = post.Visibility
	}
	moodRating := req.MoodRating
	if moodRating == nil {
		moodRating = post.MoodRating
	}

	// Update
	if err := models.UpdatePost(h.DB, postID, content, title, moodRating, visibility); err != nil {
		http.Error(w, "failed to update post", http.StatusInternalServerError)
		return
	}

	// Get updated post with details
	postDetails, err := models.GetPostWithDetails(h.DB, postID)
	if err != nil {
		http.Error(w, "failed to fetch updated post", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, postDetails)
}

// DELETE /posts/{id} — delete a post (media cascade deletes automatically)
func (h *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	postIDStr := r.PathValue("id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "invalid post id", http.StatusBadRequest)
		return
	}

	// Check ownership
	post, err := models.GetPostByID(h.DB, postID)
	if err == sql.ErrNoRows {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if post.UserID != user.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// TODO: Delete media files from S3 before deleting post
	// media, _ := models.GetMediaForPost(h.DB, postID)
	// for _, m := range media {
	//     deleteFromS3(m.FileURL)
	// }

	// Delete post (media rows cascade delete automatically)
	if err := models.DeletePost(h.DB, postID); err != nil {
		http.Error(w, "failed to delete post", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "post deleted"})
}
