package models

import (
	"database/sql"
	"time"
)

// Post represents a reflection/journal entry
type Post struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	SessionID  *int      `json:"session_id,omitempty"` // NULL for general posts
	PostType   string    `json:"post_type"`            // 'session' or 'general'
	Content    string    `json:"content,omitempty"`
	Title      string    `json:"title,omitempty"`
	MoodRating *int      `json:"mood_rating,omitempty"` // 1-5
	Visibility string    `json:"visibility"`            // 'private' or 'public'
	CreatedAt  time.Time `json:"created_at"`
}

// PostWithTags includes the post and its associated tags
type PostWithTags struct {
	Post
	Tags []string `json:"tags,omitempty"`
}

// CREATE: insert a new post
func CreatePost(db *sql.DB, userID int, sessionID *int, postType, content, title string, moodRating *int, visibility string) (Post, error) {
	var post Post

	err := db.QueryRow(
		`INSERT INTO posts (user_id, session_id, post_type, content, title, mood_rating, visibility, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		 RETURNING id, user_id, session_id, post_type, content, title, mood_rating, visibility, created_at`,
		userID, sessionID, postType, content, title, moodRating, visibility,
	).Scan(&post.ID, &post.UserID, &post.SessionID, &post.PostType, &post.Content, &post.Title, &post.MoodRating, &post.Visibility, &post.CreatedAt)

	return post, err
}

// READ: get a post by ID
func GetPostByID(db *sql.DB, postID int) (Post, error) {
	var post Post
	err := db.QueryRow(
		`SELECT id, user_id, session_id, post_type, content, title, mood_rating, visibility, created_at
		 FROM posts
		 WHERE id=$1`,
		postID,
	).Scan(&post.ID, &post.UserID, &post.SessionID, &post.PostType, &post.Content, &post.Title, &post.MoodRating, &post.Visibility, &post.CreatedAt)
	return post, err
}

// READ: get all posts for a user (their personal journal view)
// Returns posts with their tags
func GetPostsByUserID(db *sql.DB, userID int, limit, offset int) ([]PostWithTags, error) {
	rows, err := db.Query(
		`SELECT id, user_id, session_id, post_type, content, title, mood_rating, visibility, created_at
		 FROM posts
		 WHERE user_id=$1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []PostWithTags
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.UserID, &post.SessionID, &post.PostType, &post.Content, &post.Title, &post.MoodRating, &post.Visibility, &post.CreatedAt); err != nil {
			return nil, err
		}

		// Fetch tags for this post
		tags, err := GetTagsForPost(db, post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, PostWithTags{
			Post: post,
			Tags: tags,
		})
	}

	return posts, rows.Err()
}

// READ: get post linked to a specific session
func GetPostBySessionID(db *sql.DB, sessionID int) (Post, error) {
	var post Post
	err := db.QueryRow(
		`SELECT id, user_id, session_id, post_type, content, title, mood_rating, visibility, created_at
		 FROM posts
		 WHERE session_id=$1`,
		sessionID,
	).Scan(&post.ID, &post.UserID, &post.SessionID, &post.PostType, &post.Content, &post.Title, &post.MoodRating, &post.Visibility, &post.CreatedAt)
	return post, err
}

// UPDATE: update a post's content
func UpdatePost(db *sql.DB, postID int, content, title string, moodRating *int, visibility string) error {
	_, err := db.Exec(
		`UPDATE posts
		 SET content=$1, title=$2, mood_rating=$3, visibility=$4
		 WHERE id=$5`,
		content, title, moodRating, visibility, postID,
	)
	return err
}

// DELETE: remove a post by ID
func DeletePost(db *sql.DB, postID int) error {
	_, err := db.Exec(`DELETE FROM posts WHERE id=$1`, postID)
	return err
}

// Helper: get tags for a post
func GetTagsForPost(db *sql.DB, postID int) ([]string, error) {
	rows, err := db.Query(
		`SELECT t.name
		 FROM tags t
		 JOIN post_tags pt ON t.id = pt.tag_id
		 WHERE pt.post_id=$1`,
		postID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// Helper: add tags to a post
func AddTagsToPost(db *sql.DB, userID, postID int, tagNames []string) error {
	for _, tagName := range tagNames {
		// Get or create tag
		var tagID int
		err := db.QueryRow(
			`INSERT INTO tags (user_id, name)
			 VALUES ($1, $2)
			 ON CONFLICT (user_id, name) DO UPDATE SET name=EXCLUDED.name
			 RETURNING id`,
			userID, tagName,
		).Scan(&tagID)
		if err != nil {
			return err
		}

		// Link tag to post
		_, err = db.Exec(
			`INSERT INTO post_tags (post_id, tag_id)
			 VALUES ($1, $2)
			 ON CONFLICT DO NOTHING`,
			postID, tagID,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
