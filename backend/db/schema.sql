--
-- CORE ENUMS
-- These types ensure data consistency and validity in the tables below.
--

-- Defines the visibility level of a reflection post.
CREATE TYPE post_visibility AS ENUM ('private', 'public');

-- Differentiates between reflection types (Session vs. future Daily/General posts).
CREATE TYPE post_type AS ENUM ('session', 'general'); 

-- Defines the type of media attached to a post.
CREATE TYPE media_type AS ENUM ('image', 'video');


--
-- Table 1: users (User Authentication and Profile)
-- Stores core user authentication and profile data.
-- TIMESTAMPTZ (Timestamp with Time Zone) is used for accurate global tracking.
--
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,         -- Unique email used for login
    username TEXT UNIQUE,               -- Public, unique username (optional initially)
    google_id TEXT UNIQUE NOT NULL,     -- ID from Google OAuth provider
    display_name TEXT,                  -- User's preferred name for display
    picture_url TEXT,                   -- URL to the user's avatar/profile picture
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);


---

--
-- Table 2: focus_sessions (The Core Productivity Log)
-- Tracks the factual record of a user's focus time. This is the 'parent' record
-- for any session-specific reflection post.
--
CREATE TABLE IF NOT EXISTS focus_sessions (
    id SERIAL PRIMARY KEY,
    
    -- Foreign key to the user who ran the session.
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    
    -- Store the calculated duration for fast querying and stat generation.
    duration_minutes INT NOT NULL,
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Index to quickly query all sessions for a user, usually ordered by time
CREATE INDEX idx_sessions_user_id_time ON focus_sessions(user_id, start_time DESC);


---

--
-- Table 3: posts (User Reflections/Journal Entries)
-- Stores both session-specific and general reflections.
--
CREATE TABLE IF NOT EXISTS posts (
    id SERIAL PRIMARY KEY,
    
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Links to a focus_sessions record. NULL if this is a 'general' reflection.
    session_id INT REFERENCES focus_sessions(id) ON DELETE CASCADE,
    
    post_type post_type NOT NULL,       -- 'session' or 'general' (for daily reflection, etc.)
    content TEXT,                       -- The main reflection text (e.g., "What went well?")
    title TEXT,                         -- Optional title
    
    -- Reflection-specific fields
    mood_rating SMALLINT CHECK (mood_rating >= 1 AND mood_rating <= 5), -- 1 (low) to 5 (high)
    visibility post_visibility NOT NULL DEFAULT 'private',             -- Controls who can see the post
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    -- Note: updated_at column is optional, removed the trigger logic as requested.
);

-- Index for querying all posts by a user
CREATE INDEX idx_posts_user_id ON posts(user_id);
-- Index for quickly finding a reflection attached to a specific session
CREATE INDEX idx_posts_session_id ON posts(session_id);


---

--
-- Table 4: tags (Master List of Unique Tags per User)
-- Centralized storage for tag names, enabling easy sorting and renaming.
--
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    
    -- Tag ownership is tied to the user.
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    name TEXT NOT NULL,                 -- The tag name (e.g., 'deepwork')
    
    -- Constraint: Ensures a user cannot create two tags with the exact same name.
    UNIQUE(user_id, name)
);

-- Index for quickly retrieving all tags belonging to a user
CREATE INDEX idx_tags_user_id ON tags(user_id);


---

--
-- Table 5: post_tags (Junction Table for Many-to-Many Tagging)
-- Links a post to one or more tags, enabling powerful filtering.
--
CREATE TABLE IF NOT EXISTS post_tags (
    post_id INT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id INT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    
    -- Composite primary key to prevent duplicate tag associations on a single post.
    PRIMARY KEY (post_id, tag_id)
);

-- Index to efficiently find all posts for a given tag
CREATE INDEX idx_post_tags_tag_id ON post_tags(tag_id);


---

--
-- Table 6: post_media (Media Attachments)
-- Stores links to external media (images/videos), not the files themselves.
-- Media files are typically stored in services like S3 or GCS.
--
CREATE TABLE IF NOT EXISTS post_media (
    id SERIAL PRIMARY KEY,
    
    post_id INT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE, -- Redundant for safety
    
    media_type media_type NOT NULL,     -- 'image' or 'video'
    
    file_url TEXT NOT NULL,             -- The public URL where the file is hosted
    
    -- Controls the display order (e.g., image 0, 1, 2). App logic enforces the '3 per post' limit.
    position SMALLINT NOT NULL DEFAULT 0, 

    original_filename TEXT,             -- The file name provided by the user
    
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Index to retrieve all media attachments for a single post
CREATE INDEX idx_post_media_post_id ON post_media(post_id);