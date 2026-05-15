package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCommentRepository_Create(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRepository(db)

	roleID := insertRole(t, "role_c1", false)
	userID := insertUser(t, roleID, "user_c1")
	channelID := insertChannel(t, userID, "channel_c1")
	videoID := insertVideo(t, channelID, "video_c1")

	comment := &domain.Comment{UserID: userID, VideoID: videoID, Content: "hello", CreatedAt: time.Now()}
	err := repo.Create(context.Background(), comment)
	require.NoError(t, err)
	require.NotZero(t, comment.ID)
}

func TestCommentRepository_GetByID(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRepository(db)

	roleID := insertRole(t, "role_c2", false)
	userID := insertUser(t, roleID, "user_c2")
	channelID := insertChannel(t, userID, "channel_c2")
	videoID := insertVideo(t, channelID, "video_c2")
	commentID := insertComment(t, userID, videoID, "hi")

	comment, err := repo.GetByID(context.Background(), commentID)
	require.NoError(t, err)
	require.Equal(t, commentID, comment.ID)
	require.NotEmpty(t, comment.Username)
}

func TestCommentRepository_ListByVideo(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRepository(db)

	roleID := insertRole(t, "role_c3", false)
	userID := insertUser(t, roleID, "user_c3")
	channelID := insertChannel(t, userID, "channel_c3")
	videoID := insertVideo(t, channelID, "video_c3")
	_ = insertComment(t, userID, videoID, "hi")

	comments, err := repo.ListByVideo(context.Background(), videoID, 10, 0)
	require.NoError(t, err)
	require.Len(t, comments, 1)
}

func TestCommentRepository_Update(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRepository(db)

	roleID := insertRole(t, "role_c4", false)
	userID := insertUser(t, roleID, "user_c4")
	channelID := insertChannel(t, userID, "channel_c4")
	videoID := insertVideo(t, channelID, "video_c4")
	commentID := insertComment(t, userID, videoID, "old")

	comment, err := repo.GetByID(context.Background(), commentID)
	require.NoError(t, err)
	comment.Content = "new"

	err = repo.Update(context.Background(), comment)
	require.NoError(t, err)

	updated, err := repo.GetByID(context.Background(), commentID)
	require.NoError(t, err)
	require.Equal(t, "new", updated.Content)
}

func TestCommentRepository_Delete(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRepository(db)

	roleID := insertRole(t, "role_c5", false)
	userID := insertUser(t, roleID, "user_c5")
	channelID := insertChannel(t, userID, "channel_c5")
	videoID := insertVideo(t, channelID, "video_c5")
	commentID := insertComment(t, userID, videoID, "old")

	err := repo.Delete(context.Background(), commentID)
	require.NoError(t, err)

	comment, err := repo.GetByID(context.Background(), commentID)
	require.NoError(t, err)
	require.Nil(t, comment)
}

func TestCommentRepository_CountByVideo(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRepository(db)

	roleID := insertRole(t, "role_c6", false)
	userID := insertUser(t, roleID, "user_c6")
	channelID := insertChannel(t, userID, "channel_c6")
	videoID := insertVideo(t, channelID, "video_c6")
	_ = insertComment(t, userID, videoID, "one")

	count, err := repo.CountByVideo(context.Background(), videoID)
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}
