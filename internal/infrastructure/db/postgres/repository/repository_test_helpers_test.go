package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	dsn := os.Getenv("TEST_PG_DSN")
	if dsn == "" {
		os.Exit(m.Run())
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to connect to postgres:", err)
		os.Exit(1)
	}
	if err := applyMigrations(db); err != nil {
		fmt.Fprintln(os.Stderr, "failed to apply migrations:", err)
		os.Exit(1)
	}

	testDB = db
	os.Exit(m.Run())
}

func testDBOrSkip(t *testing.T) *gorm.DB {
	if testDB == nil {
		t.Skip("TEST_PG_DSN is not set")
	}
	return testDB
}

func applyMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.FromSlash("migrations/postgres/*_up.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if _, err := sqlDB.Exec(string(content)); err != nil {
			return fmt.Errorf("apply %s: %w", file, err)
		}
	}
	return nil
}

func resetDB(t *testing.T) {
	db := testDBOrSkip(t)

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	_, err = sqlDB.Exec(`
TRUNCATE TABLE
	community_post_comments,
	community_posts,
	comment_ratings,
	video_ratings,
	playlist_items,
	subscriptions,
	comments,
	viewings,
	playlists,
	videos,
	channels,
	users,
	roles
RESTART IDENTITY CASCADE;`)
	if err != nil {
		t.Fatalf("truncate tables: %v", err)
	}

	_, err = sqlDB.Exec(`
INSERT INTO roles (name, is_default)
VALUES ('admin', false), ('moderator', false), ('user', true)
ON CONFLICT (name) DO NOTHING;`)
	if err != nil {
		t.Fatalf("seed roles: %v", err)
	}
}

func insertRole(t *testing.T, name string, isDefault bool) int {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	var id int
	if err := sqlDB.QueryRow("INSERT INTO roles (name, is_default) VALUES ($1, $2) RETURNING id", name, isDefault).Scan(&id); err != nil {
		t.Fatalf("insert role: %v", err)
	}
	return id
}

func insertUser(t *testing.T, roleID int, username string) int {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	email := fmt.Sprintf("%s@example.com", strings.ToLower(username))
	var id int
	if err := sqlDB.QueryRow(`
INSERT INTO users (role_id, username, email, password_hash, is_active, notifications_enabled, created_at, updated_at)
VALUES ($1, $2, $3, $4, true, true, $5, $6)
RETURNING id`, roleID, username, email, "hash", time.Now(), time.Now()).Scan(&id); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func insertChannel(t *testing.T, userID int, name string) int {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	var id int
	if err := sqlDB.QueryRow(`
INSERT INTO channels (user_id, name, description, created_at)
VALUES ($1, $2, $3, $4)
RETURNING id`, userID, name, "desc", time.Now()).Scan(&id); err != nil {
		t.Fatalf("insert channel: %v", err)
	}
	return id
}

func insertVideo(t *testing.T, channelID int, title string) int {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	var id int
	if err := sqlDB.QueryRow(`
INSERT INTO videos (channel_id, title, description, filepath, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id`, channelID, title, "desc", "file", time.Now()).Scan(&id); err != nil {
		t.Fatalf("insert video: %v", err)
	}
	return id
}

func insertPlaylist(t *testing.T, channelID int, name string) int {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	var id int
	if err := sqlDB.QueryRow(`
INSERT INTO playlists (channel_id, name, description, created_at)
VALUES ($1, $2, $3, $4)
RETURNING id`, channelID, name, "desc", time.Now()).Scan(&id); err != nil {
		t.Fatalf("insert playlist: %v", err)
	}
	return id
}

func insertPlaylistItem(t *testing.T, playlistID, videoID, number int) {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	if _, err := sqlDB.Exec(`
INSERT INTO playlist_items (playlist_id, video_id, number, added_at)
VALUES ($1, $2, $3, $4)`, playlistID, videoID, number, time.Now()); err != nil {
		t.Fatalf("insert playlist item: %v", err)
	}
}

func insertViewing(t *testing.T, userID, videoID int) {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	if _, err := sqlDB.Exec(`
INSERT INTO viewings (user_id, video_id, watched_at)
VALUES ($1, $2, $3)`, userID, videoID, time.Now()); err != nil {
		t.Fatalf("insert viewing: %v", err)
	}
}

func insertComment(t *testing.T, userID, videoID int, content string) int {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	var id int
	if err := sqlDB.QueryRow(`
INSERT INTO comments (user_id, video_id, content, created_at)
VALUES ($1, $2, $3, $4)
RETURNING id`, userID, videoID, content, time.Now()).Scan(&id); err != nil {
		t.Fatalf("insert comment: %v", err)
	}
	return id
}

func insertSubscription(t *testing.T, userID, channelID int, newVideos int) {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	if _, err := sqlDB.Exec(`
INSERT INTO subscriptions (user_id, channel_id, new_videos_count, subscribed_at)
VALUES ($1, $2, $3, $4)`, userID, channelID, newVideos, time.Now()); err != nil {
		t.Fatalf("insert subscription: %v", err)
	}
}

func insertVideoRating(t *testing.T, userID, videoID int, liked bool) {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	if _, err := sqlDB.Exec(`
INSERT INTO video_ratings (user_id, video_id, liked, rated_at)
VALUES ($1, $2, $3, $4)`, userID, videoID, liked, time.Now()); err != nil {
		t.Fatalf("insert video rating: %v", err)
	}
}

func insertCommentRating(t *testing.T, userID, commentID int, liked bool) {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	if _, err := sqlDB.Exec(`
INSERT INTO comment_ratings (user_id, comment_id, liked, rated_at)
VALUES ($1, $2, $3, $4)`, userID, commentID, liked, time.Now()); err != nil {
		t.Fatalf("insert comment rating: %v", err)
	}
}

func insertCommunityPost(t *testing.T, channelID, userID int, content string) int {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	var id int
	if err := sqlDB.QueryRow(`
INSERT INTO community_posts (channel_id, user_id, content, created_at)
VALUES ($1, $2, $3, $4)
RETURNING id`, channelID, userID, content, time.Now()).Scan(&id); err != nil {
		t.Fatalf("insert community post: %v", err)
	}
	return id
}

func insertCommunityComment(t *testing.T, postID, userID int, content string) int {
	db := testDBOrSkip(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}

	var id int
	if err := sqlDB.QueryRow(`
INSERT INTO community_post_comments (post_id, user_id, content, created_at)
VALUES ($1, $2, $3, $4)
RETURNING id`, postID, userID, content, time.Now()).Scan(&id); err != nil {
		t.Fatalf("insert community comment: %v", err)
	}
	return id
}

func ensureTx(t *testing.T, db *gorm.DB) *sql.DB {
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	return sqlDB
}
