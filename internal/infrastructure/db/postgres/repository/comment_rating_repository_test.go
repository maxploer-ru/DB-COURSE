package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCommentRatingRepository_Create(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRatingRepository(db)

	roleID := insertRole(t, "role_cr1", false)
	userID := insertUser(t, roleID, "user_cr1")
	channelID := insertChannel(t, userID, "channel_cr1")
	videoID := insertVideo(t, channelID, "video_cr1")
	commentID := insertComment(t, userID, videoID, "c")

	err := repo.Create(context.Background(), &domain.CommentRating{UserID: userID, CommentID: commentID, Liked: true})
	require.NoError(t, err)
}

func TestCommentRatingRepository_Update(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRatingRepository(db)

	roleID := insertRole(t, "role_cr2", false)
	userID := insertUser(t, roleID, "user_cr2")
	channelID := insertChannel(t, userID, "channel_cr2")
	videoID := insertVideo(t, channelID, "video_cr2")
	commentID := insertComment(t, userID, videoID, "c")
	insertCommentRating(t, userID, commentID, true)

	updated := &domain.CommentRating{UserID: userID, CommentID: commentID, Liked: false}
	err := repo.Update(context.Background(), updated)
	require.NoError(t, err)
}

func TestCommentRatingRepository_Delete(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRatingRepository(db)

	roleID := insertRole(t, "role_cr3", false)
	userID := insertUser(t, roleID, "user_cr3")
	channelID := insertChannel(t, userID, "channel_cr3")
	videoID := insertVideo(t, channelID, "video_cr3")
	commentID := insertComment(t, userID, videoID, "c")
	insertCommentRating(t, userID, commentID, true)

	err := repo.Delete(context.Background(), userID, commentID)
	require.NoError(t, err)
}

func TestCommentRatingRepository_GetByUserAndComment(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRatingRepository(db)

	roleID := insertRole(t, "role_cr4", false)
	userID := insertUser(t, roleID, "user_cr4")
	channelID := insertChannel(t, userID, "channel_cr4")
	videoID := insertVideo(t, channelID, "video_cr4")
	commentID := insertComment(t, userID, videoID, "c")
	insertCommentRating(t, userID, commentID, true)

	rating, err := repo.GetByUserAndComment(context.Background(), userID, commentID)
	require.NoError(t, err)
	require.NotNil(t, rating)
	require.True(t, rating.Liked)
}

func TestCommentRatingRepository_GetStats(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewCommentRatingRepository(db)

	roleID := insertRole(t, "role_cr5", false)
	userID := insertUser(t, roleID, "user_cr5")
	channelID := insertChannel(t, userID, "channel_cr5")
	videoID := insertVideo(t, channelID, "video_cr5")
	commentID := insertComment(t, userID, videoID, "c")

	userID2 := insertUser(t, roleID, "user_cr6")
	insertCommentRating(t, userID, commentID, true)
	insertCommentRating(t, userID2, commentID, false)

	likes, dislikes, err := repo.GetStats(context.Background(), commentID)
	require.NoError(t, err)
	require.Equal(t, int64(1), likes)
	require.Equal(t, int64(1), dislikes)
}
