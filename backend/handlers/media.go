package handlers

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"tomo/backend/middleware"
	"tomo/backend/models"
	"tomo/backend/utils"
)

type MediaHandler struct {
	DB *sql.DB
}

const (
	MaxFileSize      = 10 << 20 // 10 MB
	MaxMediaPerPost  = 3
	AllowedImageExts = ".jpg,.jpeg,.png,.gif,.webp"
	AllowedVideoExts = ".mp4,.mov,.avi,.webm"
)

type UploadMediaRequest struct {
	PostID           int    `json:"post_id"`
	MediaType        string `json:"media_type"` // 'image' or 'video'
	FileURL          string `json:"file_url"`   // URL after uploading to S3/R2
	OriginalFilename string `json:"original_filename,omitempty"`
}

// POST /posts/{id}/media — upload media to a post
// NOTE: This expects the file to already be uploaded to S3/R2
// and you're just registering the URL in the database
func (h *MediaHandler) AddMediaToPost(w http.ResponseWriter, r *http.Request) {
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

	// Verify post exists and belongs to user
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

	// Check current media count
	count, err := models.CountMediaForPost(h.DB, postID)
	if err != nil {
		http.Error(w, "failed to count media", http.StatusInternalServerError)
		return
	}
	if count >= MaxMediaPerPost {
		http.Error(w, "maximum 3 media items per post", http.StatusBadRequest)
		return
	}

	var req UploadMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate media type
	if req.MediaType != "image" && req.MediaType != "video" {
		http.Error(w, "media_type must be 'image' or 'video'", http.StatusBadRequest)
		return
	}

	// Validate file URL
	if req.FileURL == "" {
		http.Error(w, "file_url is required", http.StatusBadRequest)
		return
	}

	// Add media to database
	media, err := models.AddMediaToPost(h.DB, postID, user.UserID, req.MediaType, req.FileURL, req.OriginalFilename, count)
	if err != nil {
		http.Error(w, "failed to add media", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, media)
}

// POST /posts/{id}/media/upload — upload file directly (multipart form)
// This is an alternative endpoint that handles the actual file upload
func (h *MediaHandler) UploadMediaFile(w http.ResponseWriter, r *http.Request) {
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

	// Verify post ownership
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

	// Check media count
	count, err := models.CountMediaForPost(h.DB, postID)
	if err != nil {
		http.Error(w, "failed to count media", http.StatusInternalServerError)
		return
	}
	if count >= MaxMediaPerPost {
		http.Error(w, "maximum 3 media items per post", http.StatusBadRequest)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		http.Error(w, "file too large (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	mediaType := ""
	if strings.Contains(AllowedImageExts, ext) {
		mediaType = "image"
	} else if strings.Contains(AllowedVideoExts, ext) {
		mediaType = "video"
	} else {
		http.Error(w, "unsupported file type", http.StatusBadRequest)
		return
	}

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}

	fileURL, err := utils.UploadToS3(r.Context(), fileBytes, header.Filename, mediaType)
	if err != nil {
		http.Error(w, "failed to upload media", http.StatusInternalServerError)
		return
	}

	// Save to database
	media, err := models.AddMediaToPost(h.DB, postID, user.UserID, mediaType, fileURL, header.Filename, count)
	if err != nil {
		http.Error(w, "failed to save media", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, media)
}

// GET /posts/{id}/media — get all media for a post
func (h *MediaHandler) GetPostMedia(w http.ResponseWriter, r *http.Request) {
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

	// Verify post access
	post, err := models.GetPostByID(h.DB, postID)
	if err == sql.ErrNoRows {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	// Check visibility
	if post.Visibility == "private" && post.UserID != user.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	media, err := models.GetMediaForPost(h.DB, postID)
	if err != nil {
		http.Error(w, "failed to fetch media", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"media": media,
		"count": len(media),
	})
}

// DELETE /media/{id} — delete a media item
func (h *MediaHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserClaims)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	mediaIDStr := r.PathValue("id")
	mediaID, err := strconv.Atoi(mediaIDStr)
	if err != nil {
		http.Error(w, "invalid media id", http.StatusBadRequest)
		return
	}

	// Get media and verify ownership
	media, err := models.GetMediaByID(h.DB, mediaID)
	if err == sql.ErrNoRows {
		http.Error(w, "media not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	if media.UserID != user.UserID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// post handler deletes from s3!

	if err := models.DeleteMedia(h.DB, mediaID); err != nil {
		http.Error(w, "failed to delete media", http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]string{"message": "media deleted"})
}
