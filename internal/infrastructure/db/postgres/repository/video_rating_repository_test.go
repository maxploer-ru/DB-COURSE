package repository

import (
	"ZVideo/internal/domain"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVideoRatingRepository_CreateAndGet(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRatingRepository(db)

	roleID := insertRole(t, "role_vrate_a", false)
	userID := insertUser(t, roleID, "user_vrate_a")
	channelID := insertChannel(t, userID, "channel_vrate_a")
	videoID := insertVideo(t, channelID, "video_vrate_a")

	rating := &domain.VideoRating{UserID: userID, VideoID: videoID, Liked: true}
	err := repo.Create(context.Background(), rating)
	require.NoError(t, err)

	loaded, err := repo.GetByUserAndVideo(context.Background(), userID, videoID)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.True(t, loaded.Liked)
}

func TestVideoRatingRepository_CreateDuplicate(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRatingRepository(db)

	roleID := insertRole(t, "role_vrate_b", false)
	userID := insertUser(t, roleID, "user_vrate_b")
	channelID := insertChannel(t, userID, "channel_vrate_b")
	videoID := insertVideo(t, channelID, "video_vrate_b")

	rating := &domain.VideoRating{UserID: userID, VideoID: videoID, Liked: true}
	err := repo.Create(context.Background(), rating)
	require.NoError(t, err)

	err = repo.Create(context.Background(), rating)
	require.ErrorIs(t, err, domain.ErrAlreadyRated)
}

func TestVideoRatingRepository_Update(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRatingRepository(db)

	roleID := insertRole(t, "role_vrate_c", false)
	userID := insertUser(t, roleID, "user_vrate_c")
	channelID := insertChannel(t, userID, "channel_vrate_c")
	videoID := insertVideo(t, channelID, "video_vrate_c")

	rating := &domain.VideoRating{UserID: userID, VideoID: videoID, Liked: true}
	require.NoError(t, repo.Create(context.Background(), rating))

	rating.Liked = false
	require.NoError(t, repo.Update(context.Background(), rating))

	loaded, err := repo.GetByUserAndVideo(context.Background(), userID, videoID)
	require.NoError(t, err)
	require.NotNil(t, loaded)
	require.False(t, loaded.Liked)
}

func TestVideoRatingRepository_Delete(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRatingRepository(db)

	roleID := insertRole(t, "role_vrate_d", false)
	userID := insertUser(t, roleID, "user_vrate_d")
	channelID := insertChannel(t, userID, "channel_vrate_d")
	videoID := insertVideo(t, channelID, "video_vrate_d")

	rating := &domain.VideoRating{UserID: userID, VideoID: videoID, Liked: true}
	require.NoError(t, repo.Create(context.Background(), rating))

	err := repo.Delete(context.Background(), userID, videoID)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), userID, videoID)
	require.ErrorIs(t, err, domain.ErrRatingNotFound)
}

func TestVideoRatingRepository_GetStats(t *testing.T) {
	resetDB(t)
	db := testDBOrSkip(t)
	repo := NewVideoRatingRepository(db)

	roleID := insertRole(t, "role_vrate_e", false)
	userID := insertUser(t, roleID, "user_vrate_e")
	channelID := insertChannel(t, userID, "channel_vrate_e")
	videoID := insertVideo(t, channelID, "video_vrate_e")

	insertVideoRating(t, userID, videoID, true)
	insertVideoRating(t, userID, videoID, false)

	likes, dislikes, err := repo.GetStats(context.Background(), videoID)
	require.NoError(t, err)
	require.Equal(t, 1, likes)
	require.Equal(t, 1, dislikes)
}
