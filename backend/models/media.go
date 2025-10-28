package models

import (
	"database/sql"
	"time"
)

// PostMedia represents a media attachment (image/video) on a post
type PostMedia struct {
	ID               int       `json:"id"`
	PostID           int       `json:"post_id"`
	UserID           int       `json:"user_id"`
	MediaType        string    `json:"media_type"` // 'image' or 'video'
	FileURL          string    `json:"file_url"`
	Position         int       `json:"position"`
	OriginalFilename string    `json:"original_filename,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// CREATE: Add media to a post
func AddMediaToPost(db *sql.DB, postID, userID int, mediaType, fileURL, originalFilename string, position int) (PostMedia, error) {
	var media PostMedia

	err := db.QueryRow(
		`INSERT INTO post_media (post_id, user_id, media_type, file_url, original_filename, position, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, NOW())
		 RETURNING id, post_id, user_id, media_type, file_url, position, original_filename, created_at`,
		postID, userID, mediaType, fileURL, originalFilename, position,
	).Scan(&media.ID, &media.PostID, &media.UserID, &media.MediaType, &media.FileURL, &media.Position, &media.OriginalFilename, &media.CreatedAt)

	return media, err
}

// READ: Get all media for a post (ordered by position)
func GetMediaForPost(db *sql.DB, postID int) ([]PostMedia, error) {
	rows, err := db.Query(
		`SELECT id, post_id, user_id, media_type, file_url, position, original_filename, created_at
		 FROM post_media
		 WHERE post_id=$1
		 ORDER BY position ASC`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []PostMedia
	for rows.Next() {
		var media PostMedia
		if err := rows.Scan(&media.ID, &media.PostID, &media.UserID, &media.MediaType, &media.FileURL, &media.Position, &media.OriginalFilename, &media.CreatedAt); err != nil {
			return nil, err
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, rows.Err()
}

// READ: Count media items for a post
func CountMediaForPost(db *sql.DB, postID int) (int, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM post_media WHERE post_id=$1`,
		postID,
	).Scan(&count)
	return count, err
}

// DELETE: Remove media by ID
func DeleteMedia(db *sql.DB, mediaID int) error {
	_, err := db.Exec(`DELETE FROM post_media WHERE id=$1`, mediaID)
	return err
}

// DELETE: Remove all media for a post (called when post is deleted)
func DeleteMediaForPost(db *sql.DB, postID int) error {
	_, err := db.Exec(`DELETE FROM post_media WHERE post_id=$1`, postID)
	return err
}

// Helper: Get media by ID
func GetMediaByID(db *sql.DB, mediaID int) (PostMedia, error) {
	var media PostMedia
	err := db.QueryRow(
		`SELECT id, post_id, user_id, media_type, file_url, position, original_filename, created_at
		 FROM post_media
		 WHERE id=$1`,
		mediaID,
	).Scan(&media.ID, &media.PostID, &media.UserID, &media.MediaType, &media.FileURL, &media.Position, &media.OriginalFilename, &media.CreatedAt)
	return media, err
}
